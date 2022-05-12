package nixconfig

import (
	"path/filepath"
)

const (
	NIX_PRELUDE      = "nix_prelude"
	NIX_REPOSITORIES = "nix_repositories"
)

// NixLanguageConfig configuration for language extension.
type NixLanguageConfig struct {
	Parent *NixLanguageConfig

	NixPrelude      string
	NixRepositories map[string]string
	Wsmode          bool
}

// NewChild creates a new child Config. It inherits desired values from the
// current Config and sets itself as the parent to the child.
func (c *NixLanguageConfig) NewChild() *NixLanguageConfig {
	return &NixLanguageConfig{
		Parent:          c,
		NixPrelude:      c.NixPrelude,
		NixRepositories: c.NixRepositories,
		Wsmode:          c.Wsmode,
	}
}

// New creates a new Config.
func New() *NixLanguageConfig {
	return &NixLanguageConfig{
		NixPrelude:      "",
		NixRepositories: make(map[string]string),
		Wsmode:          false,
	}
}

// NixLanguageConfigs is an extension of map[string]*Config.
// Aids in quicker access to method for finding package
type NixLanguageConfigs map[string]*NixLanguageConfig

// FindPackageParent returns the parent Config for the given Bazel package.
func (c *NixLanguageConfigs) FindPackageParent(pkg string) *NixLanguageConfig {
	dir := filepath.Dir(pkg)
	if dir == "." {
		dir = ""
	}

	if parent, exists := (*c)[dir]; exists {
		return parent
	}
	return nil
}
