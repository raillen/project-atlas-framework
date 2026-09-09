package packages

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

type Lock struct {
	Version  int        `json:"version"`
	Packages []Resolved `json:"packages"`
}
type Resolved struct {
	ID       string `json:"id"`
	Version  string `json:"version"`
	Checksum string `json:"checksum"`
	Type     string `json:"type"`
}

func NewLock(packages []Resolved) Lock {
	sort.Slice(packages, func(i, j int) bool { return packages[i].ID < packages[j].ID })
	return Lock{Version: 1, Packages: packages}
}
func Checksum(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
