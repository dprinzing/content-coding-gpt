package main

import (
	"content-coding-gpt/pkg/data"
	"content-coding-gpt/pkg/openai"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
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

	// Prompt Command
	promptCmd := &cobra.Command{
		Use:   "prompt <promptFile>",
		Short: "Chat complete a test prompt",
		Long:  "Chat complete a test prompt from a specified file",
		Args:  cobra.ExactArgs(1),
		RunE:  chatPrompt,
	}
	promptCmd.Flags().BoolP("raw", "r", false, "Raw OpenAI Response?")
	promptCmd.Flags().BoolP("verbose", "v", false, "Verbose output?")
	promptCmd.Flags().BoolP("system", "s", false, "Include system prompt?")
	promptCmd.Flags().IntP("max-tokens", "t", 0, "Maximum number of tokens to generate")
	promptCmd.Flags().Float32P("temperature", "T", 0.2, "Temperature for sampling")
	chatCmd.AddCommand(promptCmd)

	// Random Command
	randomCmd := &cobra.Command{
		Use:   "random <essayType>",
		Short: "Chat complete a random prompt",
		Long:  "Chat complete a random prompt of the specified type from data/original/essays.csv",
		Args:  cobra.ExactArgs(1),
		RunE:  chatRandom,
	}
	randomCmd.Flags().BoolP("raw", "r", false, "Raw OpenAI Response?")
	randomCmd.Flags().BoolP("reverse", "R", false, "Extract the score from the end of the response?")
	randomCmd.Flags().IntP("max-tokens", "t", 0, "Maximum number of tokens to generate")
	randomCmd.Flags().Float32P("temperature", "T", 0.2, "Temperature for sampling")
	randomCmd.Flags().IntP("id", "i", 0, "Essay ID (okay, not random :)")
	randomCmd.Flags().StringP("prompt", "p", "", "Prompt template text file")
	chatCmd.AddCommand(randomCmd)

	// Essay Command
	essayCmd := &cobra.Command{
		Use:   "essay <essayType> <csvFile>",
		Short: "Chat complete a specific essay type",
		Long:  "Chat complete a specific essay type from data/original/essays.csv",
		Args:  cobra.ExactArgs(2),
		RunE:  chatEssay,
	}
	essayCmd.Flags().BoolP("reverse", "R", false, "Extract the score from the end of the response?")
	essayCmd.Flags().IntP("max-tokens", "t", 0, "Maximum number of tokens to generate")
	essayCmd.Flags().Float32P("temperature", "T", 0.2, "Temperature for sampling")
	essayCmd.Flags().IntP("batch-size", "b", 10, "Batch size for concurrent requests")
	essayCmd.Flags().StringP("prompt", "p", "", "Prompt template text file")
	chatCmd.AddCommand(essayCmd)
}

// chatPrompt processes completions for a specified prompt.
func chatPrompt(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	raw, _ := cmd.Flags().GetBool("raw")
	verbose, _ := cmd.Flags().GetBool("verbose")
	system, _ := cmd.Flags().GetBool("system")
	maxTokens, _ := cmd.Flags().GetInt("max-tokens")
	temperature, _ := cmd.Flags().GetFloat32("temperature")
	promptFile := args[0]

	// Read the prompt file:
	f, err := os.Open(promptFile)
	if err != nil {
		return fmt.Errorf("error opening prompt file %s: %w", promptFile, err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("error reading prompt file %s: %w", promptFile, err)
	}
	prompt := string(b)

	// Generate the chat request:
	var messages []openai.Message
	if system {
		messages = []openai.Message{data.SystemMessage}
	}
	messages = append(messages, openai.Message{Role: openai.USER, Content: prompt})
	request := openai.ChatRequest{
		Model:       "gpt-3.5-turbo",
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	// Output the request:
	if raw || verbose {
		b, _ := json.MarshalIndent(request, "", "  ")
		fmt.Println(string(b))
	} else {
		if system {
			fmt.Println("--------------------\nSystem:")
			fmt.Println(data.SystemMessage.Content)
		}
		fmt.Println("--------------------\nUser:")
		fmt.Println(prompt)
	}

	// Raw response?
	if raw {
		response, e := apiClient.ChatCompletionRaw(ctx, request)
		if e != nil {
			return e
		}
		fmt.Print(string(response))
		return nil
	}

	// Chat complete the prompt:
	response, err := apiClient.ChatCompletion(ctx, request)
	if err != nil {
		return err
	}
	content, err := response.FirstMessageContent()
	if err != nil {
		return err
	}

	// Output the response:
	if verbose {
		b, _ := json.MarshalIndent(response, "", "  ")
		fmt.Println(string(b))
	} else {
		fmt.Println("--------------------\nAssistant:")
		fmt.Println(content)
		fmt.Println("--------------------")
	}
	return nil
}

// chatRandom chat completes a random prompt of the selected type.
func chatRandom(cmd *cobra.Command, args []string) error {
	startTime := time.Now()
	ctx := context.Background()
	raw, _ := cmd.Flags().GetBool("raw")
	reverse, _ := cmd.Flags().GetBool("reverse")
	maxTokens, _ := cmd.Flags().GetInt("max-tokens")
	temperature, _ := cmd.Flags().GetFloat32("temperature")
	id, _ := cmd.Flags().GetInt("id")
	promptFile, _ := cmd.Flags().GetString("prompt")
	essayType := args[0]
	if !data.ValidEssayType(essayType) {
		return fmt.Errorf("essay type %s is not one of: %s", essayType, strings.Join(data.EssayTypes, ", "))
	}

	// Select an essay for a chat request:
	var essay data.EssayRecord
	var err error
	if id > 0 {
		essay, err = data.ReadEssayRecord(id)
	} else {
		essay, err = data.RandomEssayRecord()
	}
	if err != nil {
		return err
	}

	// Generate the chat request:
	var request openai.ChatRequest
	if promptFile != "" {
		request, err = essay.ChatRequestTemplate(essayType, temperature, maxTokens, promptFile)
		if err != nil {
			return err
		}
	} else {
		request = essay.ChatRequest(essayType, temperature, maxTokens)
	}

	// Raw response?
	if raw {
		// Echo the Request
		jsonReq, e := json.MarshalIndent(request, "", "  ")
		if e != nil {
			return fmt.Errorf("error marshalling JSON request: %w", e)
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
	response, err := apiClient.ChatCompletion(ctx, request)
	if err != nil {
		return err
	}

	// Extract the score:
	duration := time.Since(startTime).Milliseconds()
	score, err := data.NewEssayScore(essay, essayType, response, reverse, duration)
	results := data.EssayCompletion{
		Request:  request,
		Response: response,
		Score:    score,
	}
	if err != nil {
		results.ErrMsg = err.Error()
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
	startTime := time.Now()
	ctx := context.Background()
	reverse, _ := cmd.Flags().GetBool("reverse")
	maxTokens, _ := cmd.Flags().GetInt("max-tokens")
	temperature, _ := cmd.Flags().GetFloat32("temperature")
	batchSize, _ := cmd.Flags().GetInt("batch-size")
	promptFile, _ := cmd.Flags().GetString("prompt")
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

	// Process the essays in batches:
	var count int
	scores := make([]data.EssayScore, 0, len(essays))
	batches := data.Batch(essays, batchSize)
	for i, batch := range batches {
		batchStart := time.Now()
		// Generate the chat requests:
		chats := make([]openai.Chat, 0, len(batch))
		for _, essay := range batch {
			var request openai.ChatRequest
			if promptFile != "" {
				request, err = essay.ChatRequestTemplate(essayType, temperature, maxTokens, promptFile)
				if err != nil {
					return err // error reading the template file
				}
			} else {
				request = essay.ChatRequest(essayType, temperature, maxTokens)
			}
			chat := openai.Chat{
				ID:      strconv.Itoa(essay.ID),
				Request: request,
			}
			chats = append(chats, chat)
		}
		// Process the batch:
		results := apiClient.ChatBatch(ctx, chats)
		for _, essay := range batch {
			count++
			chat, ok := results[strconv.Itoa(essay.ID)]
			if !ok {
				fmt.Printf("%d: pid %d: no response\n", count, essay.ID)
				continue
			}
			if chat.ErrMsg != "" {
				fmt.Printf("%d: pid %d: %s", count, essay.ID, chat.ErrMsg)
				continue
			}
			score, e := data.NewEssayScore(essay, essayType, chat.Response, reverse, chat.Millis)
			if e != nil {
				fmt.Printf("%d: pid %d: %v\n", count, essay.ID, e)
				continue
			}
			scores = append(scores, score)
			fmt.Printf("%d: pid %d: %.1f %d\n", count, essay.ID, score.Score, score.Millis)
		}
		// Report batch time taken, progress, and predicted time remaining:
		batchDuration := time.Since(batchStart)
		totalDuration := time.Since(startTime)
		averageDuration := totalDuration / time.Duration(count)
		timeRemaining := time.Duration(len(essays)-count) * averageDuration
		percentComplete := float32(count) / float32(len(essays)) * 100
		fmt.Printf("batch %d: %dms (%.2f%% complete, %s remaining)\n",
			i+1, batchDuration.Milliseconds(), percentComplete, timeRemaining)
	}

	// Write the scores to the specified CSV file:
	err = data.WriteEssayScores(csvFile, scores)

	// Report the total time taken:
	fmt.Printf("completed %d essays in %s\n", len(essays), time.Since(startTime))
	return err
}
