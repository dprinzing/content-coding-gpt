package data

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	PromptSeparator = "\n\n###\n\n"
	CompletionStart = " "
	CompletionStop  = "\n"
)

// TrainingRecord provides a prompt and expected completion, used to train models.
type TrainingRecord struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
}

// CleanResponse removes extra whitespace from a response.
func CleanResponse(response string) string {
	return strings.Join(strings.Fields(response), " ")
}

// WriteTrainingFile writes a JSONL file of training records.
func WriteTrainingFile(path string, append bool, records []TrainingRecord) error {
	// Open a JSONL file writer:
	var f *os.File
	var err error
	if append {
		f, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	} else {
		f, err = os.Create(path)
	}
	if err != nil {
		return fmt.Errorf("write training file %s: %w", path, err)
	}
	defer f.Close()

	// Write the records:
	for _, record := range records {
		// Marshal the record:
		j, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("write training file %s: %w", path, err)
		}
		// Write the record:
		_, err = f.Write(j)
		if err != nil {
			return fmt.Errorf("write training file %s: %w", path, err)
		}
		// Write a newline:
		_, err = f.Write([]byte("\n"))
		if err != nil {
			return fmt.Errorf("write training file %s: %w", path, err)
		}
	}
	return nil
}

// PrepareTrainingFile prepares a JSONL file of training records from the specified CSV file.
func PrepareTrainingFile(csvPath string, jsonPath string, appendFile bool) error {
	// Identify the CSV file type:
	fileType, err := IdentifyCSVFile(csvPath)
	if err != nil {
		return fmt.Errorf("prepare training file %s: %w", csvPath, err)
	}
	// Generate the training records:
	var records []TrainingRecord
	if fileType == "humility" {
		recs, e := ReadHumilityRecords(csvPath)
		if e != nil {
			return fmt.Errorf("prepare training file %s: %w", csvPath, e)
		}
		for _, r := range recs {
			records = append(records, r.PlainTrainingRecord())
		}
	} else if fileType == "spiritual" {
		recs, e := ReadSpiritualRecords(csvPath)
		if e != nil {
			return fmt.Errorf("prepare training file %s: %w", csvPath, e)
		}
		for _, r := range recs {
			records = append(records, r.PlainTrainingRecord())
		}
	} else {
		return fmt.Errorf("prepare training file %s: unexpected file type %s", csvPath, fileType)
	}
	// Write the training records:
	err = WriteTrainingFile(jsonPath, appendFile, records)
	if err != nil {
		return fmt.Errorf("prepare training file %s: %w", jsonPath, err)
	}
	return nil
}
