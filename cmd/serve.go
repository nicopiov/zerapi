package cmd

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/nicopiov/zerapi/internal/api"
	"github.com/nicopiov/zerapi/internal/loader"
	"github.com/nicopiov/zerapi/internal/store"
	"github.com/nicopiov/zerapi/internal/util"
)

func serve(args []string) error {
	if wantsHelp(args) {
		printServeHelp()
		return nil
	}

	flags := flag.NewFlagSet("serve", flag.ContinueOnError)

	port := flags.Int("port", 8080, "port to listen on")
	host := flags.String("host", "localhost", "host to listen to")
	readonly := flags.Bool("readonly", false, "block write requests")
	watch := flags.Bool("watch", false, "reload the source file when it changes")
	cors := flags.Bool("cors", false, "enable CORS headers for browser clients")
	delayValue := flags.String("delay", "", "delay every response, for example 500ms or 2s")

	flags.IntVar(port, "p", 8080, "port to listen on")

	if err := flags.Parse(args); err != nil {
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

	addr := fmt.Sprintf("%s:%d", *host, *port)
	url := fmt.Sprintf("http://%s", addr)

	data := store.New(result.Resources)
	if *watch {
		go watchFile(file, data)
	}

	var delay time.Duration
	if *delayValue != "" {
		parsedDelay, err := time.ParseDuration(*delayValue)
		if err != nil || parsedDelay < 0 {
			return fmt.Errorf("invalid delay %q: use a duration like 500ms or 2s", *delayValue)
		}
		delay = parsedDelay
	}

	handler := api.NewHandler(data, api.Options{Readonly: *readonly})
	if delay > 0 {
		handler = api.WithDelay(handler, delay)
	}
	if *cors {
		handler = api.WithCORS(handler)
	}
	handler = api.WithLogging(handler, os.Stdout)

	printStartup(url, file, result.Resources, *readonly, *watch, *cors, delay)

	return http.ListenAndServe(addr, handler)
}

func printStartup(url string, file string, resources []loader.Resource, readonly bool, watch bool, cors bool, delay time.Duration) {
	fmt.Printf("Zerapi running at %s\n", util.Info(url))
	fmt.Printf("%s %s\n", util.Success("Loaded"), file)

	if readonly {
		fmt.Printf("%s readonly\n", util.Warn("Mode:"))
		fmt.Println(util.Muted("Writes are disabled"))
	}

	if watch {
		fmt.Printf("%s enabled\n", util.Info("Watch:"))
	}

	if cors {
		fmt.Printf("%s enabled\n", util.Info("CORS:"))
	}

	if delay > 0 {
		fmt.Printf("%s %s\n", util.Info("Delay:"), delay)
	}

	fmt.Println()
	fmt.Println("Resources:")

	totalRecords := 0
	for _, resource := range resources {
		totalRecords += len(resource.Records)

		fmt.Printf("  %s /%s\n", util.Info("GET   "), resource.Name)
		fmt.Printf("  %s /%s/{id}\n", util.Info("GET   "), resource.Name)
		if readonly {
			fmt.Printf("  %s /%s %s\n", util.Muted("POST  "), resource.Name, util.Warn("(readonly)"))
			fmt.Printf("  %s /%s/{id} %s\n", util.Muted("PUT   "), resource.Name, util.Warn("(readonly)"))
			fmt.Printf("  %s /%s/{id} %s\n", util.Muted("PATCH "), resource.Name, util.Warn("(readonly)"))
			fmt.Printf("  %s /%s/{id} %s\n", util.Muted("DELETE"), resource.Name, util.Warn("(readonly)"))
			continue
		}
		fmt.Printf("  %s /%s\n", util.Info("POST  "), resource.Name)
		fmt.Printf("  %s /%s/{id}\n", util.Info("PUT   "), resource.Name)
		fmt.Printf("  %s /%s/{id}\n", util.Info("PATCH "), resource.Name)
		fmt.Printf("  %s /%s/{id}\n", util.Info("DELETE"), resource.Name)
	}

	fmt.Println()
	fmt.Printf("%s %d records\n", util.Success("Loaded"), totalRecords)
}

func watchFile(file string, data *store.Store) {
	lastModTime := modTime(file)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		currentModTime := modTime(file)
		if currentModTime.IsZero() || !currentModTime.After(lastModTime) {
			continue
		}

		lastModTime = currentModTime

		result, err := loader.Load(file)
		if err != nil {
			fmt.Printf("%s reload failed: %v\n", util.Warn("Watch:"), err)
			continue
		}

		data.Reload(result.Resources)
		fmt.Printf("%s reloaded %s\n", util.Success("Watch:"), file)
	}
}

func modTime(file string) time.Time {
	info, err := os.Stat(file)
	if err != nil {
		return time.Time{}
	}

	return info.ModTime()
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
  --delay       Delay every response, for example 500ms or 2s`)
}
