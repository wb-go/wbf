// Package cleanenvport provides a unified way to load and validate application configuration
// from a file (YAML/JSON/TOML) using cleanenv and validator.
package cleanenvport

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

var (
	// ErrConfigPathNotSet is returned when neither --config flag nor CONFIG_PATH env var is set.
	ErrConfigPathNotSet = errors.New("config path not set")
	// ErrConfigFileNotFound is returned when the config file does not exist.
	ErrConfigFileNotFound = errors.New("config file not found")
	// ErrConfigValidation is returned when the config structure fails validation.
	ErrConfigValidation = errors.New("config validation failed")
)

// Load reads configuration from a file path specified via --config flag or CONFIG_PATH
// environment variable, then validates the config structure using validator tags.
// Returns an error if the path is not set, the file doesn't exist, or validation fails.
func Load(cfg any) error {
	path := fetchConfigPath()
	if path == "" {
		return fmt.Errorf("%w (use --config flag or CONFIG_PATH env)", ErrConfigPathNotSet)
	}
	return LoadPath(path, cfg)
}

// LoadPath loads and validates configuration from the given file path.
// It checks that the file exists, reads it using cleanenv, and validates
// the resulting struct using go-playground/validator.
// Returns a descriptive error on any failure.
func LoadPath(configPath string, cfg any) error {
	validate := validator.New()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", ErrConfigFileNotFound, configPath)
	}

	if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("%w: %w", ErrConfigValidation, formatValidationError(err))
	}

	return nil
}

// formatValidationError converts validator.ValidationErrors into a human-readable string.
// Each field error is formatted as "FieldName=value (tag)", and all are joined with "; ".
func formatValidationError(err error) error {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		var msgs []string
		for _, ve := range validationErrs {
			msgs = append(msgs, fmt.Sprintf("%s=%v (%s)", ve.Field(), ve.Value(), ve.Tag()))
		}
		return fmt.Errorf("%w: %s", ErrConfigValidation, strings.Join(msgs, "; "))
	}
	return fmt.Errorf("%w: %w", ErrConfigValidation, err)
}

// fetchConfigPath retrieves the configuration file path from the --config command-line flag.
// If the flag is not defined, it registers and parses it.
// If the flag is empty, it falls back to the CONFIG_PATH environment variable.
// Returns the resolved path or an empty string if neither is set.
func fetchConfigPath() string {
	var path string
	if f := flag.Lookup("config"); f != nil {
		path = f.Value.String()
	} else {
		flag.StringVar(&path, "config", "", "path to config file")
		flag.Parse()
	}
	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}
	return path
}
