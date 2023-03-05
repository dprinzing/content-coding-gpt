package main

import (
	"content-coding-gpt/pkg/data"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// initCompleteCmd initializes the complete commands.
func initCompleteCmd(root *cobra.Command) {
	// Complete Command
	completeCmd := &cobra.Command{
		Use:   "complete",
		Short: "Complete a prompt",
		Long:  "Complete a prompt",
	}
	root.AddCommand(completeCmd)

	// Random Command
	randomCmd := &cobra.Command{
		Use:   "random <essayType> <modelID>",
		Short: "Complete a random prompt",
		Long:  "Complete a random prompt of the specified type from data/original/essays.csv",
		RunE:  completeRandom,
	}
	randomCmd.Flags().BoolP("raw", "r", false, "Raw OpenAI Response?")
	randomCmd.Flags().IntP("max-tokens", "t", 6, "Maximum number of tokens to generate")
	completeCmd.AddCommand(randomCmd)

	// Essay Command
	essayCmd := &cobra.Command{
		Use:   "essay <essayType> <modelID> <csvFile>",
		Short: "Complete a specific essay type",
		Long:  "Complete a specific essay type from data/original/essays.csv",
		RunE:  completeEssay,
	}
	essayCmd.Flags().IntP("max-tokens", "t", 6, "Maximum number of tokens to generate")
	completeCmd.AddCommand(essayCmd)
}

// completeRandom completes a random prompt.
func completeRandom(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	raw, _ := cmd.Flags().GetBool("raw")
	maxTokens, _ := cmd.Flags().GetInt("max-tokens")
	essayType := args[0]
	if !data.ValidEssayType(essayType) {
		return fmt.Errorf("essay type %s is not one of: %s", essayType, strings.Join(data.EssayTypes, ", "))
	}
	modelID := args[1]

	// Select a random essay:
	essay, err := data.RandomEssayRecord()
	if err != nil {
		return err
	}

	// Create the request:
	request := essay.PlainCompletionRequest(essayType, modelID, maxTokens)
	jsonReq, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON request: %w", err)
	}

	// Raw response?
	if raw {
		body, err := apiClient.CreateCompletionRaw(ctx, request)
		if err != nil {
			return err
		}
		fmt.Println(string(jsonReq))
		fmt.Print(string(body))
		return nil
	}

	// Complete the prompt:
	completion, err := apiClient.CreateCompletion(ctx, request)
	if err != nil {
		return err
	}
	fmt.Println(string(jsonReq))
	j, err := json.MarshalIndent(completion, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON completion: %w", err)
	}
	fmt.Println(string(j))

	// Humility?
	if data.IsHumility(essayType) {
		r, err := data.NewHumilityRecord(essay, essayType, completion)
		if err != nil {
			return err
		}
		j, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling JSON HumilityRecord: %w", err)
		}
		fmt.Println(string(j))
	}

	// Spiritual?
	if data.IsSpiritual(essayType) {
		r, err := data.NewSpiritualRecord(essay, essayType, completion)
		if err != nil {
			return err
		}
		j, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling JSON SpiritualRecord: %w", err)
		}
		fmt.Println(string(j))
	}

	return nil
}

// completeEssay processes completions for all essays of a specified type for
// specified model. The output is is placed in the specified CSV file.
func completeEssay(cmd *cobra.Command, args []string) error {
	startTime := time.Now()
	ctx := context.Background()
	maxTokens, _ := cmd.Flags().GetInt("max-tokens")
	csvFile := args[2]

	// Validate the specified essay type:
	essayType := args[0]
	if !data.ValidEssayType(essayType) {
		return fmt.Errorf("essay type %s is not one of: %s", essayType, strings.Join(data.EssayTypes, ", "))
	}

	// Validate the specified model:
	modelID := args[1]
	_, err := apiClient.ReadModel(ctx, modelID)
	if err != nil {
		return err
	}

	// Load the essays:
	essays, err := data.ReadEssayRecords("data/original/essays.csv")
	if err != nil {
		return err
	}

	// Humility?
	if data.IsHumility(essayType) {
		records := make([]data.HumilityRecord, 0, len(essays))
		for i, essay := range essays {
			request := essay.PlainCompletionRequest(essayType, modelID, maxTokens)
			completion, err := apiClient.CreateCompletion(ctx, request)
			if err != nil {
				fmt.Printf("%d: pid %d: %v\n", i, essay.ID, err)
				continue
			}
			r, err := data.NewHumilityRecord(essay, essayType, completion)
			if err != nil {
				fmt.Printf("%d: pid %d: %v\n", i, essay.ID, err)
				continue
			}
			records = append(records, r)
			fmt.Printf("%d: pid %d: %s\n", i, essay.ID, r.Results())
		}
		err = data.WriteHumilityRecords(csvFile, records)
	}

	// Spiritual?
	if data.IsSpiritual(essayType) {
		records := make([]data.SpiritualRecord, 0, len(essays))
		for i, essay := range essays {
			request := essay.PlainCompletionRequest(essayType, modelID, maxTokens)
			completion, err := apiClient.CreateCompletion(ctx, request)
			if err != nil {
				fmt.Printf("%d: pid %d: %v\n", i, essay.ID, err)
				continue
			}
			r, err := data.NewSpiritualRecord(essay, essayType, completion)
			if err != nil {
				fmt.Printf("%d: pid %d: %v\n", i, essay.ID, err)
				continue
			}
			records = append(records, r)
			fmt.Printf("%d: pid %d: %s\n", i, essay.ID, r.Results())
		}
		err = data.WriteSpiritualRecords(csvFile, records)
	}

	// Report the time taken:
	fmt.Printf("completed %d essays in %s\n", len(essays), time.Since(startTime))
	return err
}
