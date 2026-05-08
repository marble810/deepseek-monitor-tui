package main

import (
	"encoding/json"
	"fmt"
	"os"

	"deepseekMon/api"
	"deepseekMon/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	result := api.NewScraper(cfg.PlatformToken).FetchAll()
	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result.StatusSummary())
	fmt.Println(string(out))

	if len(result.Errors) > 0 {
		os.Exit(1)
	}
}
