package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/nicopiov/zerapi/internal/loader"
	"github.com/nicopiov/zerapi/internal/store"
	"github.com/nicopiov/zerapi/internal/util"
)

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
