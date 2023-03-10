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
		Short:   "gpt: OpenAI GPT content coding",
		Long:    "gpt is a command line tool for content coding with OpenAI GPT models",
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
	initChatCmd(rootCmd)
	initCompleteCmd(rootCmd)
	initFileCmd(rootCmd)
	initModelCmd(rootCmd)
	initTuneCmd(rootCmd)

	// Execute the specified command:
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
