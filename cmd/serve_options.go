package cmd

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nicopiov/zerapi/internal/config"
)

type serveOptions struct {
	host       string
	port       int
	readonly   bool
	watch      bool
	cors       bool
	delayValue string
}

func resolveServeOptions(args []string) (serveOptions, *flag.FlagSet, error) {
	defaults := config.ServeConfig{
		Host: "localhost",
		Port: 8080,
	}

	if path := configPathFromArgs(args); path != "" {
		loaded, err := config.Load(path)
		if err != nil {
			return serveOptions{}, nil, err
		}

		defaults = mergeServeConfig(defaults, *loaded)
	}

	defaultPort, err := envInt("ZERAPI_PORT", defaults.Port)
	if err != nil {
		return serveOptions{}, nil, err
	}
	defaultReadonly, err := envBool("ZERAPI_READONLY", defaults.Readonly)
	if err != nil {
		return serveOptions{}, nil, err
	}
	defaultWatch, err := envBool("ZERAPI_WATCH", defaults.Watch)
	if err != nil {
		return serveOptions{}, nil, err
	}
	defaultCORS, err := envBool("ZERAPI_CORS", defaults.CORS)
	if err != nil {
		return serveOptions{}, nil, err
	}
	defaultDelay := envString("ZERAPI_DELAY", defaults.Delay)

	flags := flag.NewFlagSet("serve", flag.ContinueOnError)
	_ = flags.String("config", "", "load serve options from a JSON or YAML config file")
	port := flags.Int("port", defaultPort, "port to listen on")
	host := flags.String("host", envString("ZERAPI_HOST", defaults.Host), "host to listen to")
	readonly := flags.Bool("readonly", defaultReadonly, "block write requests")
	watch := flags.Bool("watch", defaultWatch, "reload the source file when it changes")
	cors := flags.Bool("cors", defaultCORS, "enable CORS headers for browser clients")
	delayValue := flags.String("delay", defaultDelay, "delay every response, for example 500ms or 2s")

	flags.IntVar(port, "p", defaultPort, "port to listen on")

	if err := flags.Parse(args); err != nil {
		return serveOptions{}, nil, err
	}

	return serveOptions{
		host:       *host,
		port:       *port,
		readonly:   *readonly,
		watch:      *watch,
		cors:       *cors,
		delayValue: *delayValue,
	}, flags, nil
}

func parseDelay(value string) (time.Duration, error) {
	if value == "" {
		return 0, nil
	}

	delay, err := time.ParseDuration(value)
	if err != nil || delay < 0 {
		return 0, fmt.Errorf("invalid delay %q: use a duration like 500ms or 2s", value)
	}

	return delay, nil
}

func envString(name string, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}

func envBool(name string, fallback bool) (bool, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean", name)
	}
	return parsed, nil
}

func envInt(name string, fallback int) (int, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", name)
	}
	return parsed, nil
}

func configPathFromArgs(args []string) string {
	for i, arg := range args {
		if arg == "--config" && i+1 < len(args) {
			return args[i+1]
		}

		if strings.HasPrefix(arg, "--config=") {
			return strings.TrimPrefix(arg, "--config=")
		}
	}

	return ""
}

func mergeServeConfig(base config.ServeConfig, override config.ServeConfig) config.ServeConfig {
	if override.Host != "" {
		base.Host = override.Host
	}

	if override.Port != 0 {
		base.Port = override.Port
	}

	if override.Readonly {
		base.Readonly = true
	}

	if override.Watch {
		base.Watch = true
	}

	if override.CORS {
		base.CORS = true
	}

	if override.Delay != "" {
		base.Delay = override.Delay
	}

	return base
}
