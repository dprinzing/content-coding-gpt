package main

import (
	"content-coding-gpt/pkg/openai"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// initTuneCmd initializes the fine-tune commands.
func initTuneCmd(root *cobra.Command) {
	// Tune Command
	tuneCmd := &cobra.Command{
		Use:   "tune",
		Short: "Manage fine-tuned models",
		Long:  "Manage fine-tuned models.",
	}
	root.AddCommand(tuneCmd)

	// List Command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List fine-tuned models",
		Long:  "List metadata of available fine-tuned models.",
		RunE:  listTunes,
	}
	listCmd.Flags().BoolP("verbose", "v", false, "Verbose? (full JSON)")
	listCmd.Flags().BoolP("raw", "r", false, "Raw OpenAI Response?")
	tuneCmd.AddCommand(listCmd)

	// Read Command
	readCmd := &cobra.Command{
		Use:   "read <tuneID> [tuneID]...",
		Short: "Read specified fine-tuned model(s)",
		Long:  "Read the metadata about one or more fine-tuned models, specified by ID.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  readTune,
	}
	readCmd.Flags().BoolP("raw", "r", false, "Raw OpenAI Response?")
	tuneCmd.AddCommand(readCmd)

	// Events Command
	eventsCmd := &cobra.Command{
		Use:   "events <tuneID>",
		Short: "List events for a fine-tuned model",
		Long:  "List the event history for a specified fine-tuned model.",
		Args:  cobra.ExactArgs(1),
		RunE:  listTuneEvents,
	}
	eventsCmd.Flags().BoolP("verbose", "v", false, "Verbose? (full JSON)")
	eventsCmd.Flags().BoolP("raw", "r", false, "Raw OpenAI Response?")
	tuneCmd.AddCommand(eventsCmd)

	// Create Command
	createCmd := &cobra.Command{
		Use:   "create <trainingFileID> [validationFileID]",
		Short: "Create a fine-tuned model",
		Long:  "Create a fine-tuned model from the provided training file ID.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  createTune,
	}
	createCmd.Flags().BoolP("raw", "r", false, "Raw OpenAI Response?")
	createCmd.Flags().StringP("base", "b", "curie", "Base model (default: curie)")
	createCmd.Flags().StringP("suffix", "s", "", "Name suffix of the fine-tuned model")
	tuneCmd.AddCommand(createCmd)

	// Cancel Command
	cancelCmd := &cobra.Command{
		Use:   "cancel <tuneID> [tuneID]...",
		Short: "Cancel specified fine-tuned model(s)",
		Long:  "Cancel one or more fine-tuned models, specified by ID.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  cancelTune,
	}
	tuneCmd.AddCommand(cancelCmd)

	// Delete Command
	deleteCmd := &cobra.Command{
		Use:   "delete <tuneID> [tuneID]...",
		Short: "Delete specified fine-tuned model(s)",
		Long:  "Delete one or more fine-tuned models, specified by ID.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  deleteTune,
	}
	tuneCmd.AddCommand(deleteCmd)
}

// listTunes lists the fine-tuned models.
func listTunes(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the raw OpenAI response?
	raw, err := cmd.Flags().GetBool("raw")
	if err != nil {
		return err
	}
	if raw {
		body, e := apiClient.ListFineTunesRaw(ctx)
		if e != nil {
			return e
		}
		fmt.Println(string(body))
		return nil
	}

	// Retrieve the fine-tuned models.
	tunes, err := apiClient.ListFineTunes(ctx)
	if err != nil {
		return err
	}

	// Print the fine-tuned models.
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return err
	}
	if verbose {
		j, err := json.MarshalIndent(tunes, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling FineTune JSON: %w", err)
		}
		fmt.Println(string(j))
	} else {
		for _, tune := range tunes {
			fmt.Println(tune.ID, tune.Status, tune.FineTunedModel)
		}
	}
	return nil
}

// readTune reads the fine-tuned models.
func readTune(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the raw OpenAI response?
	raw, err := cmd.Flags().GetBool("raw")
	if err != nil {
		return err
	}
	if raw {
		for _, id := range args {
			body, err := apiClient.ReadFineTuneRaw(ctx, id)
			if err != nil {
				return err
			}
			fmt.Println(string(body))
		}
	} else {
		for _, id := range args {
			tune, err := apiClient.ReadFineTune(ctx, id)
			if err != nil {
				return err
			}
			j, err := json.MarshalIndent(tune, "", "  ")
			if err != nil {
				return fmt.Errorf("error marshalling FineTune JSON: %w", err)
			}
			fmt.Println(string(j))
		}
	}
	return nil
}

// listTuneEvents lists the events for a fine-tuned model.
func listTuneEvents(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the raw OpenAI response?
	raw, err := cmd.Flags().GetBool("raw")
	if err != nil {
		return err
	}
	if raw {
		body, e := apiClient.ListFineTuneEventsRaw(ctx, args[0])
		if e != nil {
			return e
		}
		fmt.Println(string(body))
		return nil
	}

	// Retrieve the events.
	events, err := apiClient.ListFineTuneEvents(ctx, args[0])
	if err != nil {
		return err
	}

	// Print the events.
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return err
	}
	if verbose {
		j, err := json.MarshalIndent(events, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling Events JSON: %w", err)
		}
		fmt.Println(string(j))
	} else {
		for _, event := range events {
			t := time.Unix(event.CreatedAt, 0)
			fmt.Println(t, event.Level, event.Message)
		}
	}
	return nil
}

// createTune creates a fine-tuned model.
func createTune(cmd *cobra.Command, args []string) error {
	// Gather request parameters
	ctx := context.Background()
	base := cmd.Flag("base").Value.String()
	suffix := cmd.Flag("suffix").Value.String()
	trainingFileID := args[0]
	validationFileID := ""
	if len(args) > 1 {
		validationFileID = args[1]
	}
	raw, err := cmd.Flags().GetBool("raw")
	if err != nil {
		return err
	}

	// Validate the base model.
	if !apiClient.ValidModel(ctx, base) {
		return fmt.Errorf("invalid base model: %s", base)
	}

	// Validate the training file ID.
	_, err = apiClient.ReadFile(ctx, trainingFileID)
	if err != nil {
		return fmt.Errorf("invalid training file ID %s: %w", trainingFileID, err)
	}

	// Validate the validation file ID.
	if validationFileID != "" {
		_, err := apiClient.ReadFile(ctx, validationFileID)
		if err != nil {
			return fmt.Errorf("invalid validation file ID %s: %w", validationFileID, err)
		}
	}

	// Create the fine-tuned model, returning the raw response if requested.
	req := openai.FineTuneRequest{
		TrainingFileID:   trainingFileID,
		ValidationFileID: validationFileID,
		Model:            base,
		Suffix:           suffix,
	}
	if raw {
		body, err := apiClient.CreateFineTuneRaw(ctx, req)
		if err != nil {
			return err
		}
		fmt.Println(string(body))
	} else {
		tune, err := apiClient.CreateFineTune(ctx, req)
		if err != nil {
			return err
		}
		j, err := json.MarshalIndent(tune, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling FineTune JSON: %w", err)
		}
		fmt.Println(string(j))
	}
	return nil
}

// cancelTune cancels a fine-tuned model job in progress.
func cancelTune(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	for _, id := range args {
		tune, err := apiClient.CancelFineTune(ctx, id)
		if err != nil {
			return err
		}
		j, err := json.MarshalIndent(tune, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling FineTune JSON: %w", err)
		}
		fmt.Println(string(j))
	}
	return nil
}

// deleteTune deletes specified fine-tuned model(s).
func deleteTune(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	for _, id := range args {
		err := apiClient.DeleteFineTune(ctx, id)
		if err != nil {
			return err
		}
	}
	return nil
}
