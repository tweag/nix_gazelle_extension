package nixconfig

import "path/filepath"

const (
	NixPrelude      = "nix_prelude"
	NixRepositories = "nix_repositories"
)

// Configs is an extension of map[string]*Config. It provides finding methods
// on top of the mapping.
type Configs map[string]*Config

// Config configuration for language extension.
type Config struct {
	parent *Config

	nixPrelude      string
	nixRepositories map[string]string
	wsmode          bool
}

// NewChild creates a new child Config. It inherits desired values from the
// current Config and sets itself as the parent to the child.
func (c *Config) NewChild() *Config {
	return &Config{
		parent:          c,
		nixPrelude:      c.nixPrelude,
		nixRepositories: c.nixRepositories,
		wsmode:          c.wsmode,
	}
}

// New creates a new Config.
func New() *Config {
	return &Config{
		nixPrelude:      "",
		nixRepositories: make(map[string]string),
		wsmode:          false,
	}
}

// ParentForPackage returns the parent Config for the given Bazel package.
func (c *Configs) ParentForPackage(pkg string) *Config {
	dir := filepath.Dir(pkg)
	if dir == "." {
		dir = ""
	}
	parent := (map[string]*Config)(*c)[dir]
	return parent
}

func (c *Config) SetNixPrelude(filename string) {
	c.nixPrelude = filename
}

func (c Config) NixPrelude() string {
	return c.nixPrelude
}

func (c *Config) SetNixRepositories(repositories map[string]string) {
	c.nixRepositories = repositories
}

func (c Config) NixRepositories() map[string]string {
	return c.nixRepositories
}

func (c *Config) SetWsMode(wsmode bool) {
	c.wsmode = wsmode
}

func (c Config) WsMode() bool {
	return c.wsmode
}
