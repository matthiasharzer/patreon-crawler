package main

import (
	"fmt"
	"os"

	"github.com/MatthiasHarzer/patreon-crawler/cmd/crawl"
	"github.com/MatthiasHarzer/patreon-crawler/cmd/version"

	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Use:           "patreon-crawler",
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCommand.AddCommand(version.Command)
	rootCommand.AddCommand(crawl.Command)
}

func main() {
	err := rootCommand.Execute()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
