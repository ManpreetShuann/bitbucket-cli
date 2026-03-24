package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// SaveCredentials stores a token for a profile in credentials.yaml with 0600 permissions.
func SaveCredentials(credsPath, profile, token string) error {
	if err := os.MkdirAll(filepath.Dir(credsPath), 0755); err != nil {
		return err
	}

	v := viper.New()
	v.SetConfigFile(credsPath)
	v.SetConfigType("yaml")
	_ = v.ReadInConfig() // ignore: file may not exist yet

	v.Set(fmt.Sprintf("profiles.%s.token", profile), token)

	if err := v.WriteConfigAs(credsPath); err != nil {
		return err
	}
	return os.Chmod(credsPath, 0600)
}

// LoadToken reads a token for a profile from credentials.yaml.
func LoadToken(credsPath, profile string) (string, error) {
	v := viper.New()
	v.SetConfigFile(credsPath)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return "", nil
	}
	return v.GetString(fmt.Sprintf("profiles.%s.token", profile)), nil
}

// RemoveCredentials removes a profile's token from credentials.yaml.
func RemoveCredentials(credsPath, profile string) error {
	v := viper.New()
	v.SetConfigFile(credsPath)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil
	}

	v.Set(fmt.Sprintf("profiles.%s.token", profile), "")
	return v.WriteConfigAs(credsPath)
}
