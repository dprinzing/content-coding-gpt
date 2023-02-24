package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// initModelCmd initializes the model commands.
func initModelCmd(root *cobra.Command) {
	// Model Command
	modelCmd := &cobra.Command{
		Use:   "model",
		Short: "Manage models",
		Long:  "Manage models",
	}
	root.AddCommand(modelCmd)

	// List Command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List models",
		Long:  "List available models",
		RunE:  listModels,
	}
	listCmd.Flags().BoolP("verbose", "v", false, "Verbose? (full JSON)")
	listCmd.Flags().BoolP("raw", "r", false, "Raw OpenAI Response?")
	modelCmd.AddCommand(listCmd)

	// Read Command
	readCmd := &cobra.Command{
		Use:   "read <modelID> [modelID]...",
		Short: "Read specified model(s)",
		Long:  "Read the details about one or more models, specified by ID.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  readModel,
	}
	readCmd.Flags().BoolP("raw", "r", false, "Raw OpenAI Response?")
	modelCmd.AddCommand(readCmd)
}

// listModels lists the available models.
func listModels(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the raw JSON response:
	raw, _ := cmd.Flags().GetBool("raw")
	if raw {
		body, err := apiClient.ListModelsRaw(ctx)
		if err != nil {
			return err
		}
		fmt.Println(string(body))
		return nil
	}

	// Retrieve the models:
	models, err := apiClient.ListModels(ctx)
	if err != nil {
		return err
	}

	// Display either full JSON or just the IDs:
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		j, err := json.MarshalIndent(models, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling JSON models: %w", err)
		}
		fmt.Println(string(j))
	} else {
		for _, model := range models {
			fmt.Println(model.ID)
		}
	}
	return nil
}

// readModel reads the details about specified model(s).
func readModel(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the raw JSON response:
	raw, _ := cmd.Flags().GetBool("raw")
	if raw {
		for _, modelID := range args {
			body, err := apiClient.ReadModelRaw(ctx, modelID)
			if err != nil {
				return err
			}
			fmt.Println(string(body))
		}
		return nil
	}

	// Retrieve the model(s):
	for _, modelID := range args {
		model, err := apiClient.ReadModel(ctx, modelID)
		if err != nil {
			return err
		}
		j, err := json.MarshalIndent(model, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling JSON model: %w", err)
		}
		fmt.Println(string(j))
	}
	return nil
}
