package main

import (
	"content-coding-gpt/pkg/openai"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	promptSeparator = "\n\n###\n\n"
	completionStart = " "
	completionStop  = "\n"
)

// ReadCSVFile reads a CSV file and returns the records.
func ReadCSVFile(path string, skipHeader bool) ([][]string, error) {
	// Open a CSV file reader:
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read csv file %s: %w", path, err)
	}
	defer f.Close()
	r := csv.NewReader(f)

	// Skip the header if specified:
	if skipHeader {
		_, err := r.Read()
		if err != nil {
			return nil, fmt.Errorf("read csv file %s: %w", path, err)
		}
	}

	// Read and return the records:
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv file %s: %w", path, err)
	}
	return records, nil
}

// ReadJSONLFile reads a JSONL file and returns the file bytes.
func ReadJSONLFile(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read jsonl file %s: %w", path, err)
	}
	return b, nil
}

// WriteJSONLFile writes a JSONL file.
func WriteJSONLFile(path string, append bool, records []openai.TrainingRecord) error {
	// Open a JSONL file writer:
	var f *os.File
	var err error
	if append {
		f, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	} else {
		f, err = os.Create(path)
	}
	if err != nil {
		return fmt.Errorf("write jsonl file %s: %w", path, err)
	}
	defer f.Close()

	// Write the records:
	for _, record := range records {
		// Marshal the record:
		jsonl, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("write jsonl file %s: %w", path, err)
		}
		// Write the record:
		_, err = f.Write(jsonl)
		if err != nil {
			return fmt.Errorf("write jsonl file %s: %w", path, err)
		}
		// Write a newline:
		_, err = f.Write([]byte("\n"))
		if err != nil {
			return fmt.Errorf("write jsonl file %s: %w", path, err)
		}
	}
	return nil
}

// PrepareTrainingRecords prepares the training records.
func PrepareTrainingRecords(input [][]string) []openai.TrainingRecord {
	// Initialize the training records:
	records := make([]openai.TrainingRecord, 0, len(input))
	// Iterate over the input records, skipping empty records:
	for _, record := range input {
		if len(record) > 1 && record[0] != "" && record[1] != "" {
			// Initialize the training record:
			prompt := strings.Join(strings.Fields(record[0]), " ") + promptSeparator
			completion := completionStart + strings.Join(record[1:], " ") + completionStop
			records = append(records, openai.TrainingRecord{
				Prompt:     prompt,
				Completion: completion,
			})
		}
	}
	return records
}
