package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Profiles []Profile `toml:"profile"`
}

type Profile struct {
	Name       string `toml:"name"`
	GitName    string `toml:"git_name"`
	GitEmail   string `toml:"git_email"`
	SigningKey string `toml:"signing_key,omitempty"`
	SSHKey     string `toml:"ssh_key,omitempty"`
}

// path follows XDG, so it is $XDG_CONFIG_HOME/ggu/config.toml or ~/.config/ggu/config.toml.
func path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "ggu", "config.toml"), nil
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
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return fmt.Errorf("create %s: %w", filepath.Dir(p), err)
	}
	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("write %s: %w", p, err)
	}
	defer func() { _ = f.Close() }()
	return toml.NewEncoder(f).Encode(cfg)
}

// ExpandHome turns a leading "~/" into the user's home directory.
func ExpandHome(path string) string {
	rest, ok := strings.CutPrefix(path, "~/")
	if !ok {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, rest)
}

// Normalized trims every field and stores home paths as "~/", so the form and
// the flags save the same input the same way.
func (p Profile) Normalized() Profile {
	return Profile{
		Name:       strings.TrimSpace(p.Name),
		GitName:    strings.TrimSpace(p.GitName),
		GitEmail:   strings.TrimSpace(p.GitEmail),
		SigningKey: strings.TrimSpace(p.SigningKey),
		SSHKey:     ContractHome(strings.TrimSpace(p.SSHKey)),
	}
}

// ContractHome turns the user's home directory back into "~/", keeping the config portable
// across machines, since shells expand "~" before ggu sees a flag value.
func ContractHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if rest, ok := strings.CutPrefix(path, home+string(filepath.Separator)); ok {
		return "~/" + rest
	}
	return path
}

// CheckKeyPath accepts only full paths ("/..." or "~/...") to an existing file,
// since relative paths would break as soon as git runs in another directory.
func CheckKeyPath(path string) error {
	if !strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "~/") {
		return errors.New("use a full path")
	}
	info, err := os.Stat(ExpandHome(path))
	switch {
	case err != nil:
		return errors.New("file not found")
	case info.IsDir():
		return errors.New("that is a folder")
	}
	return nil
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

func (c *Config) Remove(name string) bool {
	for i, p := range c.Profiles {
		if p.Name == name {
			c.Profiles = append(c.Profiles[:i], c.Profiles[i+1:]...)
			return true
		}
	}
	return false
}
