package main

import (
	"content-coding-gpt/pkg/openai"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var apiClient *openai.Client

// main is the entry point for the application.
func main() {
	// Initialize the API client:
	apiClient = openai.NewClient("", "")

	// Root Command
	rootCmd := &cobra.Command{
		Use:     "gpt",
		Short:   "gpt is a command line tool for content coding with GPT-3",
		Long:    "gpt is a command line tool for content coding with GPT-3",
		Version: "0.0.1",
	}

	// About Command
	aboutCmd := &cobra.Command{
		Use:   "about",
		Short: "Print application information",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("gpt is a command line tool for content coding with GPT-3")
			return nil
		},
	}
	rootCmd.AddCommand(aboutCmd)

	// Initialize the commands:
	initModelCmd(rootCmd)
	initFileCmd(rootCmd)
	initTuneCmd(rootCmd)
	initCompleteCmd(rootCmd)

	// Execute the specified command:
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
