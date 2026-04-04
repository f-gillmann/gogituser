package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Profiles []Profile `toml:"profile"`
}

type Profile struct {
	Name       string `toml:"name"`
	GitName    string `toml:"git_name"`
	Email      string `toml:"email"`
	SigningKey string `toml:"signing_key"`
	SSHKey     string `toml:"ssh_key"`
	SSHHost    string `toml:"ssh_host,omitempty"`
}

func path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gogituser.toml"), nil
}

func Load() (*Config, error) {
	p, err := path()
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return cfg, nil
	}
	if _, err := toml.DecodeFile(p, cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", p, err)
	}
	return cfg, nil
}

func Save(cfg *Config) error {
	p, err := path()
	if err != nil {
		return err
	}
	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("write %s: %w", p, err)
	}
	defer func() { _ = f.Close() }()
	return toml.NewEncoder(f).Encode(cfg)
}

func (c *Config) Find(name string) *Profile {
	for i := range c.Profiles {
		if c.Profiles[i].Name == name {
			return &c.Profiles[i]
		}
	}
	return nil
}

func (c *Config) Upsert(p Profile) {
	for i := range c.Profiles {
		if c.Profiles[i].Name == p.Name {
			c.Profiles[i] = p
			return
		}
	}
	c.Profiles = append(c.Profiles, p)
}

func (c *Config) Delete(name string) bool {
	for i, p := range c.Profiles {
		if p.Name == name {
			c.Profiles = append(c.Profiles[:i], c.Profiles[i+1:]...)
			return true
		}
	}
	return false
}
