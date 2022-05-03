package nixconfig

const (
	NixPrelude      = "nix_prelude"
	NixRepositories = "nix_repositories"
)

// Configs is an extension of map[string]*Config. It provides finding methods
// on top of the mapping.
type Configs map[string]*Config

// Config configuration for language extension.
type Config struct {
	nixPrelude      string
	nixRepositories map[string]string
}

// New creates a new Config.
func New() *Config {
	return &Config{
		nixPrelude:      "",
		nixRepositories: make(map[string]string),
	}
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
