package app

import (
	"context"

	"github.com/raillen/prumo/internal/project"
	"github.com/raillen/prumo/internal/protocol"
)

type VersionInfo struct {
	Version         string `json:"version"`
	ProtocolVersion string `json:"protocol_version"`
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Version(_ context.Context) VersionInfo {
	return VersionInfo{
		Version:         protocol.CLIVersion,
		ProtocolVersion: protocol.ProtocolVersion,
	}
}

type ProjectInfo struct {
	Root     string           `json:"root"`
	Manifest project.Manifest `json:"manifest"`
}

func (s *Service) ProjectRoot(_ context.Context, start string) (ProjectInfo, error) {
	root, err := project.FindRoot(start)
	if err != nil {
		return ProjectInfo{}, err
	}
	manifest, err := project.LoadManifest(root)
	if err != nil {
		return ProjectInfo{}, err
	}
	return ProjectInfo{Root: root, Manifest: manifest}, nil
}
