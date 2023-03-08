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

// initChatCmd initializes the chat commands.
func initChatCmd(root *cobra.Command) {
	// Chat Command
	chatCmd := &cobra.Command{
		Use:   "chat",
		Short: "Complete a chat prompt",
		Long:  "Complete a chat prompt",
	}
	root.AddCommand(chatCmd)

	// Random Command
	randomCmd := &cobra.Command{
		Use:   "random <essayType>",
		Short: "Chat complete a random prompt",
		Long:  "Chat complete a random prompt of the specified type from data/original/essays.csv",
		RunE:  chatRandom,
	}
	randomCmd.Flags().BoolP("raw", "r", false, "Raw OpenAI Response?")
	randomCmd.Flags().IntP("max-tokens", "t", 0, "Maximum number of tokens to generate")
	randomCmd.Flags().Float32P("temperature", "T", 0.2, "Temperature for sampling")
	chatCmd.AddCommand(randomCmd)

	// Essay Command
	essayCmd := &cobra.Command{
		Use:   "essay <essayType> <csvFile>",
		Short: "Chat complete a specific essay type",
		Long:  "Chat complete a specific essay type from data/original/essays.csv",
		RunE:  chatEssay,
	}
	essayCmd.Flags().IntP("max-tokens", "t", 0, "Maximum number of tokens to generate")
	essayCmd.Flags().Float32P("temperature", "T", 0.2, "Temperature for sampling")
	chatCmd.AddCommand(essayCmd)
}

// chatRandom chat completes a random prompt of the selected type.
func chatRandom(cmd *cobra.Command, args []string) error {
	startTime := time.Now()
	ctx := context.Background()
	raw, _ := cmd.Flags().GetBool("raw")
	maxTokens, _ := cmd.Flags().GetInt("max-tokens")
	temperature, _ := cmd.Flags().GetFloat32("temperature")
	essayType := args[0]
	if !data.ValidEssayType(essayType) {
		return fmt.Errorf("essay type %s is not one of: %s", essayType, strings.Join(data.EssayTypes, ", "))
	}

	// Select a random essay for a chat request:
	essay, err := data.RandomEssayRecord()
	if err != nil {
		return err
	}
	request := essay.ChatRequest(essayType, temperature, maxTokens)

	// Raw response?
	if raw {
		// Echo the Request
		jsonReq, err := json.MarshalIndent(request, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling JSON request: %w", err)
		}
		fmt.Println(string(jsonReq))
		// Output the Response
		body, e := apiClient.ChatCompletionRaw(ctx, request)
		if e != nil {
			return e
		}
		fmt.Print(string(body))
		return nil
	}

	// Chat complete the prompt:
	chat, err := apiClient.ChatCompletion(ctx, request)
	if err != nil {
		return err
	}

	// Extract the score:
	score, err := data.NewEssayScore(essay, essayType, chat, time.Since(startTime))
	if err != nil {
		return err
	}
	results := data.EssayCompletion{
		Request:  request,
		Response: chat,
		Score:    score,
	}
	j, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON chat completion: %w", err)
	}
	fmt.Println(string(j))
	return nil
}

// chatEssay processes completions for all essays of a specified type for
// specified model. The output is is placed in the specified CSV file.
func chatEssay(cmd *cobra.Command, args []string) error {
	batchStart := time.Now()
	ctx := context.Background()
	maxTokens, _ := cmd.Flags().GetInt("max-tokens")
	temperature, _ := cmd.Flags().GetFloat32("temperature")
	csvFile := args[1]

	// Validate the specified essay type:
	essayType := args[0]
	if !data.ValidEssayType(essayType) {
		return fmt.Errorf("essay type %s is not one of: %s", essayType, strings.Join(data.EssayTypes, ", "))
	}

	// Load the essays:
	essays, err := data.ReadEssayRecords("data/original/essays.csv")
	if err != nil {
		return err
	}

	// Process the essays:
	records := make([]data.EssayScore, 0, len(essays))
	for i, essay := range essays {
		startTime := time.Now()
		request := essay.ChatRequest(essayType, temperature, maxTokens)
		chat, e := apiClient.ChatCompletion(ctx, request)
		if e != nil {
			fmt.Printf("%d: pid %d: %v\n", i, essay.ID, e)
			continue
		}
		s, e := data.NewEssayScore(essay, essayType, chat, time.Since(startTime))
		if e != nil {
			fmt.Printf("%d: pid %d: %v\n", i, essay.ID, e)
			continue
		}
		records = append(records, s)
		fmt.Printf("%d: pid %d: %.2f %d\n", i, essay.ID, s.Score, s.Millis)
	}
	err = data.WriteEssayScores(csvFile, records)

	// Report the time taken:
	fmt.Printf("completed %d essays in %s\n", len(essays), time.Since(batchStart))
	return err
}
