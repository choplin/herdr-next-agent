package main

import (
	"context"
	"os"

	"github.com/choplin/herdr-next-agent/internal/navigation"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	direction, err := parseDirection(args)
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		return 2
	}
	cfg, err := navigation.LoadConfig(environmentOrDefault("HERDR_PLUGIN_CONFIG_DIR", ""))
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		return 1
	}
	binary := environmentOrDefault("HERDR_BIN_PATH", "herdr")
	client := navigation.NewHerdrClient(binary, navigation.OSRunner{})
	if err := navigation.Run(context.Background(), client, cfg, direction); err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		return 1
	}
	return 0
}

// ArgumentError reports an unsupported command invocation.
type ArgumentError struct{}

func (*ArgumentError) Error() string {
	return "usage: herdr-next-agent [next|previous]"
}

func parseDirection(args []string) (navigation.Direction, error) {
	if len(args) == 0 || len(args) == 1 && args[0] == "next" {
		return navigation.NextDirection, nil
	}
	if len(args) == 1 && args[0] == "previous" {
		return navigation.PreviousDirection, nil
	}
	return 0, &ArgumentError{}
}

func environmentOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
