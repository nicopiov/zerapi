package cmd

import (
	"fmt"
)

const version = "v0.4.0"

func Execute(args []string) error {
	if len(args) < 1 {
		printHelp()
		return fmt.Errorf("missing command")
	}

	switch args[0] {
	case "version", "--version", "-v":
		fmt.Printf("zerapi %s\n", version)
		return nil
	case "help", "-h", "--help":
		printHelp()
		return nil
	case "serve":
		return serve(args[1:])
	default:
		printHelp()
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func printHelp() {
	fmt.Println(`Zerapi - Instant APIs. Zero setup.

Usage:
  zerapi <command>
  
Commands:
  serve <file>   Start a local API from a data file
  version        Print the version of Zerapi

Server flags:
  --host         Host to listen on (default: localhost)
  --port, -p     Port to listen on (default: 8080)
  --readonly     Block POST, PUT, PATCH, and DELETE requests
  --watch        Reload the source file when it changes
  --cors         Enable CORS headers for browser clients
  --delay        Delay every response, for example 500ms or 2s`)
}
