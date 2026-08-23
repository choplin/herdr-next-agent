package navigation

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const configFileName = "config.toml"

var supportedStates = map[string]struct{}{
	"blocked": {},
	"done":    {},
	"idle":    {},
	"unknown": {},
	"working": {},
}

// SelectionOrder controls how matching Agents are traversed.
type SelectionOrder string

const (
	// DisplayOrder follows the order reported for Herdr's Agents sidebar.
	DisplayOrder SelectionOrder = "display"
	// WaitingOrder visits Agents with the oldest current state first.
	WaitingOrder SelectionOrder = "waiting"
)

// Config contains user-configurable navigation behavior.
type Config struct {
	States []string       `toml:"states"`
	Order  SelectionOrder `toml:"order"`
}

// ConfigReadError reports an inaccessible or malformed config file.
type ConfigReadError struct {
	Path string
	Err  error
}

func (err *ConfigReadError) Error() string {
	return "read Next Agent configuration " + err.Path + ": " + err.Err.Error()
}

// Unwrap exposes the underlying file or TOML error.
func (err *ConfigReadError) Unwrap() error {
	return err.Err
}

// InvalidConfigError reports unsupported configuration values or fields.
type InvalidConfigError struct {
	Detail string
}

func (err *InvalidConfigError) Error() string {
	return "invalid Next Agent configuration: " + err.Detail
}

// DefaultConfig returns the zero-configuration behavior.
func DefaultConfig() Config {
	return Config{States: []string{"blocked"}, Order: DisplayOrder}
}

// LoadConfig reads config.toml from the plugin config directory.
func LoadConfig(dir string) (Config, error) {
	cfg := DefaultConfig()
	if dir == "" {
		return cfg, nil
	}

	path := filepath.Join(dir, configFileName)
	metadata, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return Config{}, &ConfigReadError{Path: path, Err: err}
	}
	if undecoded := metadata.Undecoded(); len(undecoded) != 0 {
		fields := make([]string, 0, len(undecoded))
		for _, field := range undecoded {
			fields = append(fields, field.String())
		}
		return Config{}, &InvalidConfigError{Detail: "unknown field " + strings.Join(fields, ", ")}
	}

	if cfg.Order != DisplayOrder && cfg.Order != WaitingOrder {
		return Config{}, &InvalidConfigError{Detail: "order must be \"display\" or \"waiting\""}
	}
	if len(cfg.States) == 0 {
		return Config{}, &InvalidConfigError{Detail: "states must contain at least one Agent state"}
	}
	seen := make(map[string]struct{}, len(cfg.States))
	for _, state := range cfg.States {
		if _, ok := supportedStates[state]; !ok {
			return Config{}, &InvalidConfigError{
				Detail: "unsupported state " + state,
			}
		}
		if _, ok := seen[state]; ok {
			return Config{}, &InvalidConfigError{Detail: "duplicate state " + state}
		}
		seen[state] = struct{}{}
	}
	return cfg, nil
}
