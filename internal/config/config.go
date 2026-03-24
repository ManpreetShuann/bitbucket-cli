package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the resolved configuration for a CLI session.
type Config struct {
	URL            string
	Token          string
	Profile        string
	DefaultProject string
	DefaultRepo    string
}

// ConfigDir returns the config directory path.
func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "bb")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "bb")
}

// Load resolves configuration using: CLI flags > env vars > config files.
func Load(profile, configDir string) (*Config, error) {
	if configDir == "" {
		configDir = ConfigDir()
	}
	return LoadFromDir(configDir, profile)
}

// LoadFromDir loads config from a specific directory (useful for testing).
func LoadFromDir(dir, profile string) (*Config, error) {
	cfg := &Config{Profile: profile}

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(dir)
	v.ReadInConfig()

	profilePrefix := fmt.Sprintf("profiles.%s", profile)

	cfg.URL = v.GetString(profilePrefix + ".url")
	cfg.DefaultProject = v.GetString(profilePrefix + ".default-project")
	cfg.DefaultRepo = v.GetString(profilePrefix + ".default-repo")

	credsPath := filepath.Join(dir, "credentials.yaml")
	token, _ := LoadToken(credsPath, profile)
	cfg.Token = token

	if envURL := os.Getenv("BITBUCKET_URL"); envURL != "" {
		cfg.URL = envURL
	}
	if envToken := os.Getenv("BITBUCKET_TOKEN"); envToken != "" {
		cfg.Token = envToken
	}

	return cfg, nil
}

// SaveProfile writes a profile to the config file.
func SaveProfile(configDir, profile, url, defaultProject, defaultRepo string) error {
	configPath := filepath.Join(configDir, "config.yaml")
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	v.ReadInConfig()

	prefix := fmt.Sprintf("profiles.%s", profile)
	v.Set(prefix+".url", url)
	if defaultProject != "" {
		v.Set(prefix+".default-project", defaultProject)
	}
	if defaultRepo != "" {
		v.Set(prefix+".default-repo", defaultRepo)
	}
	v.Set("current-profile", profile)

	os.MkdirAll(configDir, 0755)
	return v.WriteConfigAs(configPath)
}

// ClearDefaults removes default-project and default-repo from a profile.
func ClearDefaults(configDir, profile string) error {
	configPath := filepath.Join(configDir, "config.yaml")
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")
	v.ReadInConfig()

	prefix := fmt.Sprintf("profiles.%s", profile)
	v.Set(prefix+".default-project", "")
	v.Set(prefix+".default-repo", "")

	return v.WriteConfigAs(configPath)
}
