package main

import (
	"content-coding-gpt/pkg/data"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// initFileCmd initializes the file commands.
func initFileCmd(root *cobra.Command) {
	// File Command
	fileCmd := &cobra.Command{
		Use:   "file",
		Short: "Manage files",
		Long:  "Manage files",
	}
	root.AddCommand(fileCmd)

	// List Command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List files",
		Long:  "List metadata of available files",
		RunE:  listFiles,
	}
	listCmd.Flags().BoolP("verbose", "v", false, "Verbose? (full JSON)")
	listCmd.Flags().BoolP("raw", "r", false, "Raw OpenAI Response?")
	fileCmd.AddCommand(listCmd)

	// Read Command
	readCmd := &cobra.Command{
		Use:   "read <fileID> [fileID]...",
		Short: "Read specified file(s)",
		Long:  "Read the metadata about one or more files, specified by ID.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  readFile,
	}
	readCmd.Flags().BoolP("raw", "r", false, "Raw OpenAI Response?")
	fileCmd.AddCommand(readCmd)

	// Prepare Command
	prepareCmd := &cobra.Command{
		Use:   "prepare <csvFile> <jsonlFile>",
		Short: "Prepare a JSONL file for upload",
		Long:  "Prepare a JSONL fine-tuning file for upload",
		Args:  cobra.ExactArgs(2),
		RunE:  prepareFile,
	}
	prepareCmd.Flags().BoolP("append", "a", false, "Append to existing file?")
	fileCmd.AddCommand(prepareCmd)

	// Upload Command
	uploadCmd := &cobra.Command{
		Use:   "upload <jsonlFile>",
		Short: "Upload a JSONL file",
		Long:  "Upload a JSONL fine-tuning file",
		Args:  cobra.ExactArgs(1),
		RunE:  uploadFile,
	}
	uploadCmd.Flags().StringP("purpose", "p", "fine-tune", "File Purpose")
	fileCmd.AddCommand(uploadCmd)

	// Download Command
	downloadCmd := &cobra.Command{
		Use:   "download <fileID>",
		Short: "Download a file",
		Long:  "Download a file",
		Args:  cobra.ExactArgs(1),
		RunE:  downloadFile,
	}
	downloadCmd.Flags().StringP("output", "o", "", "Output File Path")
	fileCmd.AddCommand(downloadCmd)

	// Delete Command
	deleteCmd := &cobra.Command{
		Use:   "delete <fileID> [fileID]...",
		Short: "Delete specified file(s)",
		Long:  "Delete one or more files, specified by ID.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  deleteFile,
	}
	fileCmd.AddCommand(deleteCmd)
}

// listFiles lists the available files.
func listFiles(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the raw JSON response:
	raw, _ := cmd.Flags().GetBool("raw")
	if raw {
		body, err := apiClient.ListFilesRaw(ctx)
		if err != nil {
			return err
		}
		fmt.Println(string(body))
		return nil
	}

	// Retrieve the files:
	files, err := apiClient.ListFiles(ctx)
	if err != nil {
		return err
	}

	// Display either full JSON or just the IDs:
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		j, err := json.MarshalIndent(files, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling JSON files: %w", err)
		}
		fmt.Println(string(j))
	} else {
		for _, file := range files {
			fmt.Println(file.ID, file.Purpose, file.FileName)
		}
	}
	return nil
}

// readFile reads the details about specified file(s).
func readFile(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the raw JSON response:
	raw, _ := cmd.Flags().GetBool("raw")
	if raw {
		for _, fileID := range args {
			body, err := apiClient.ReadFileRaw(ctx, fileID)
			if err != nil {
				return err
			}
			fmt.Print(string(body))
		}
		return nil
	}

	// Retrieve the file(s):
	for _, fileID := range args {
		file, err := apiClient.ReadFile(ctx, fileID)
		if err != nil {
			return err
		}
		j, err := json.MarshalIndent(file, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling JSON file: %w", err)
		}
		fmt.Println(string(j))
	}
	return nil
}

// prepareFile prepares a JSONL fine-tuning file for upload.
func prepareFile(cmd *cobra.Command, args []string) error {
	append, _ := cmd.Flags().GetBool("append")
	csvPath := args[0]
	jsonPath := args[1]
	return data.PrepareTrainingFile(csvPath, jsonPath, append)
}

// uploadFile uploads a JSONL fine-tuning file.
func uploadFile(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	purpose := cmd.Flag("purpose").Value.String()
	path := args[0]
	fileName := filepath.Base(path)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("upload file %s: %w", path, err)
	}
	file, err := apiClient.UploadFile(ctx, fileName, purpose, data)
	if err != nil {
		return err
	}
	j, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON file: %w", err)
	}
	fmt.Println(string(j))
	return nil
}

// downloadFile downloads the specified file.
func downloadFile(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	path := cmd.Flag("output").Value.String()
	fileID := args[0]
	if path == "" {
		// Read the file metadata to get the file name:
		file, err := apiClient.ReadFile(ctx, fileID)
		if err != nil {
			return fmt.Errorf("download file %s: %w", fileID, err)
		}
		path = file.FileName
	}
	// Download the file:
	data, err := apiClient.DownloadFile(ctx, fileID)
	if err != nil {
		return fmt.Errorf("download file %s: %w", fileID, err)
	}
	// Write the file to disk:
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("download file %s: %w", fileID, err)
	}
	return nil
}

// deleteFile deletes the specified file.
func deleteFile(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	for _, fileID := range args {
		err := apiClient.DeleteFile(ctx, fileID)
		if err != nil {
			return err
		}
	}
	return nil
}
