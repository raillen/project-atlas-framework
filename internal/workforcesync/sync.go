package workforcesync

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/raillen/prumo/internal/install"
	"github.com/raillen/prumo/internal/packages"
	"github.com/raillen/prumo/internal/protocol"
	"github.com/raillen/prumo/internal/resolver"
)

var (
	ErrSkillNotFound   = errors.New("skill not found in catalog or remote")
	ErrOfflineRequired = errors.New("cannot download skill: offline mode enabled and package not cached")
)

type Options struct {
	ProjectRoot   string
	HomeDir       string
	RepoRoot      string
	RemoteBaseURL string
	Version       string
	OfflineOnly   bool
	ForceRemote   bool
	TargetSkills  []string
}

type Result struct {
	ResolvedSkills  []string            `json:"resolved_skills"`
	InstalledSkills []string            `json:"installed_skills"`
	CacheHits       []string            `json:"cache_hits"`
	RemoteDownloads []string            `json:"remote_downloads"`
	LockPackages    []packages.Resolved `json:"lock_packages"`
}

type Service struct {
	HTTPClient *http.Client
}

func NewService() *Service {
	return &Service{
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *Service) Sync(opts Options) (*Result, error) {
	if opts.ProjectRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		opts.ProjectRoot = cwd
	}
	if opts.HomeDir == "" {
		home, err := install.HomeDir("")
		if err != nil {
			return nil, err
		}
		opts.HomeDir = home
	}
	if opts.Version == "" {
		opts.Version = protocol.CLIVersion
	}
	if opts.RemoteBaseURL == "" {
		opts.RemoteBaseURL = fmt.Sprintf("https://raw.githubusercontent.com/raillen/prumo/v%s", opts.Version)
	}

	// 1. Determine skills to install
	skillsToInstall := opts.TargetSkills
	if len(skillsToInstall) == 0 {
		var profile resolver.Profile
		profilePath := filepath.Join(opts.ProjectRoot, "prumo.json")
		if p, err := resolver.LoadProfile(profilePath); err == nil {
			profile = p
		} else {
			profile = resolver.Profile{Raw: map[string]any{}}
		}

		catalogRoot := opts.RepoRoot
		if catalogRoot == "" {
			// Check if we are within the prumo repository
			if _, err := os.Stat("src/prumo/resources/catalog/catalog.json"); err == nil {
				catalogRoot = "."
			}
		}

		if catalogRoot != "" {
			catalog, err := resolver.LoadCatalog(catalogRoot)
			if err == nil {
				resolution := catalog.Resolve(profile)
				skillsToInstall = resolution.Skills
			}
		}

		// Always ensure core skills are included
		coreSkills := []string{"clean-code", "cognitive-clarity", "testing-quality"}
		present := map[string]bool{}
		for _, s := range skillsToInstall {
			present[s] = true
		}
		for _, c := range coreSkills {
			if !present[c] {
				skillsToInstall = append(skillsToInstall, c)
			}
		}
	}
	sort.Strings(skillsToInstall)

	result := &Result{
		ResolvedSkills:  skillsToInstall,
		InstalledSkills: []string{},
		CacheHits:       []string{},
		RemoteDownloads: []string{},
		LockPackages:    []packages.Resolved{},
	}

	targetDir := filepath.Join(opts.ProjectRoot, ".ai", "skills")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, err
	}

	for _, skillID := range skillsToInstall {
		cacheSkillDir := filepath.Join(opts.HomeDir, "cache", "workforce", opts.Version, "skills", skillID)
		manifestPath := filepath.Join(cacheSkillDir, "manifest.json")

		inCache := false
		if !opts.ForceRemote {
			if info, err := os.Stat(manifestPath); err == nil && !info.IsDir() {
				inCache = true
			}
		}

		if inCache {
			result.CacheHits = append(result.CacheHits, skillID)
		} else {
			// Populate cache from RepoRoot if available
			populated := false
			if opts.RepoRoot != "" {
				repoSkillDir := filepath.Join(opts.RepoRoot, "src", "prumo", "resources", "workforce", "skills", skillID)
				if info, err := os.Stat(repoSkillDir); err == nil && info.IsDir() {
					_ = copyDir(repoSkillDir, cacheSkillDir)
					populated = true
				}
			}

			// If not populated from repo, attempt remote download
			if !populated {
				if opts.OfflineOnly {
					return nil, fmt.Errorf("%w: %s", ErrOfflineRequired, skillID)
				}
				if err := s.downloadSkill(opts.RemoteBaseURL, skillID, cacheSkillDir); err != nil {
					// Fallback to local fallback if exists in relative directory
					fallbackDir := filepath.Join("src", "prumo", "resources", "workforce", "skills", skillID)
					if info, statErr := os.Stat(fallbackDir); statErr == nil && info.IsDir() {
						_ = copyDir(fallbackDir, cacheSkillDir)
					} else {
						return nil, fmt.Errorf("failed to fetch skill %q: %w", skillID, err)
					}
				} else {
					result.RemoteDownloads = append(result.RemoteDownloads, skillID)
				}
			}
		}

		// Copy from cache to project .ai/skills/<skillID>
		destSkillDir := filepath.Join(targetDir, skillID)
		if err := copyDir(cacheSkillDir, destSkillDir); err != nil {
			return nil, fmt.Errorf("failed to install skill %q into project: %w", skillID, err)
		}
		result.InstalledSkills = append(result.InstalledSkills, skillID)

		// Calculate checksum of installed skill manifest
		destManifest := filepath.Join(destSkillDir, "manifest.json")
		manifestData, err := os.ReadFile(destManifest)
		var sum string
		if err == nil {
			sum = packages.Checksum(manifestData)
		} else {
			sum = "unknown"
		}

		result.LockPackages = append(result.LockPackages, packages.Resolved{
			ID:       skillID,
			Version:  opts.Version,
			Checksum: sum,
			Type:     "skill",
		})
	}

	// Update .ai/skills/manifest.json
	skillsManifest := map[string]any{
		"version": opts.Version,
		"skills":  result.InstalledSkills,
		"updated": time.Now().UTC().Format(time.RFC3339),
	}
	data, err := json.MarshalIndent(skillsManifest, "", "  ")
	if err == nil {
		_ = os.WriteFile(filepath.Join(targetDir, "manifest.json"), append(data, '\n'), 0644)
	}

	// Update prumo.lock
	lock := packages.NewLock(result.LockPackages)
	lockData, err := json.MarshalIndent(lock, "", "  ")
	if err == nil {
		_ = os.WriteFile(filepath.Join(opts.ProjectRoot, "prumo.lock"), append(lockData, '\n'), 0644)
	}

	return result, nil
}

func (s *Service) downloadSkill(baseURL, skillID, destDir string) error {
	client := s.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	files := []string{
		"manifest.json",
		"SKILL.md",
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	for _, file := range files {
		url := fmt.Sprintf("%s/src/prumo/resources/workforce/skills/%s/%s", strings.TrimRight(baseURL, "/"), skillID, file)
		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("remote responded with HTTP %d for %s", resp.StatusCode, url)
		}

		destPath := filepath.Join(destDir, file)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}
		out, err := os.Create(destPath)
		if err != nil {
			return err
		}
		_, err = io.Copy(out, resp.Body)
		out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func copyDir(source, target string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(target, rel)
		if info.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0644)
	})
}
