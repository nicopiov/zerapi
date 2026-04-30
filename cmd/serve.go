package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/nicopiov/zerapi/internal/api"
	"github.com/nicopiov/zerapi/internal/loader"
	"github.com/nicopiov/zerapi/internal/store"
)

func serve(args []string) error {
	if wantsHelp(args) {
		printServeHelp()
		return nil
	}

	options, flags, err := resolveServeOptions(args)
	if err != nil {
		return err
	}

	if flags.NArg() != 1 {
		return fmt.Errorf("usage: zerapi serve [--host host] [--port port] <file>")
	}

	file := flags.Arg(0)

	result, err := loader.Load(file)
	if err != nil {
		return fmt.Errorf("load %s: %w", file, err)
	}

	delay, err := parseDelay(options.delayValue)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", options.host, options.port)
	url := fmt.Sprintf("http://%s", addr)

	data := store.New(result.Resources)
	if options.watch {
		go watchFile(file, data)
	}

	handler := api.NewHandler(data, api.Options{Readonly: options.readonly})
	if delay > 0 {
		handler = api.WithDelay(handler, delay)
	}
	if options.cors {
		handler = api.WithCORS(handler)
	}
	handler = api.WithLogging(handler, os.Stdout)

	printStartup(url, file, result.Resources, options.readonly, options.watch, options.cors, delay)

	return http.ListenAndServe(addr, handler)
}

func wantsHelp(args []string) bool {
	for _, arg := range args {
		if arg == "help" || arg == "--help" || arg == "-h" {
			return true
		}
	}

	return false
}

func printServeHelp() {
	fmt.Println(`Zerapi serve

Usage:
  zerapi serve [flags] <file>

Flags:
  --host        Host to listen on (default: localhost)
  --port, -p    Port to listen on (default: 8080)
  --readonly    Block POST, PUT, PATCH, and DELETE requests
  --watch       Reload the source file when it changes
  --cors        Enable CORS headers for browser clients
  --delay       Delay every response, for example 500ms or 2s
  --config      Load serve options from a JSON or YAML config file`)
}
