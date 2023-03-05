package data

import (
	"content-coding-gpt/pkg/openai"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
)

// EssayTypes is a list of supported essay types.
// Note that "conflict" and "angry" are equivalent.
var EssayTypes = []string{"dream", "dejavu", "conflict", "angry", "award"}

// ValidEssayType returns true if the specified essay type is valid.
func ValidEssayType(essayType string) bool {
	for _, t := range EssayTypes {
		if t == essayType {
			return true
		}
	}
	return false
}

// IsSpiritual returns true if the specified essay type is spiritual.
func IsSpiritual(essayType string) bool {
	return essayType == "dream" || essayType == "dejavu"
}

// IsHumility returns true if the specified essay type is humility.
func IsHumility(essayType string) bool {
	return essayType == "conflict" || essayType == "angry" || essayType == "award"
}

// EssayRecord contains responses that need to be content-coded.
type EssayRecord struct {
	ID       int    `csv:"pid"`
	Dream    string `csv:"dream"`
	Dejavu   string `csv:"dejavu"`
	Conflict string `csv:"conflict"`
	Award    string `csv:"award"`
}

// SelectEssay returns the specified essay type from the EssayRecord.
func (r EssayRecord) SelectEssay(essayType string) string {
	switch essayType {
	case "dream":
		return r.Dream
	case "dejavu":
		return r.Dejavu
	case "conflict", "angry":
		return r.Conflict
	case "award":
		return r.Award
	default:
		return ""
	}
}

// PlainPrompt converts an EssayRecord to a plain prompt for the specified essay response.
func (r EssayRecord) PlainPrompt(essayType string) string {
	switch essayType {
	case "dream":
		return r.Dream + PromptSeparator
	case "dejavu":
		return r.Dejavu + PromptSeparator
	case "conflict", "angry":
		return r.Conflict + PromptSeparator
	case "award":
		return r.Award + PromptSeparator
	default:
		return ""
	}
}

// CompletionRequest converts an EssayRecord into an OpenAI CompletionRequest
// using a plain prompt. The essayType is used to determine which prompt to use
// and the model is used to determine which model to use.
func (r EssayRecord) PlainCompletionRequest(essayType string, model string, maxTokens int) openai.CompletionRequest {
	return openai.CompletionRequest{
		Model:       model,
		Prompt:      r.PlainPrompt(essayType),
		MaxTokens:   maxTokens,
		Temperature: 0.2,
		Stop:        []string{PromptSeparator},
		User:        strconv.Itoa(r.ID),
	}
}

// NewEssayRecord creates a new EssayRecord from a slice of strings,
// ostensibly read from a CSV file.
func NewEssayRecord(fields []string) (EssayRecord, error) {
	var record EssayRecord
	var err error
	if len(fields) != 5 {
		return record, errors.New("invalid number of fields")
	}
	record.ID, err = strconv.Atoi(fields[0])
	if err != nil {
		return record, fmt.Errorf("invalid pid %s: %w", fields[0], err)
	}
	record.Dream = CleanResponse(fields[1])
	if record.Dream == "" {
		return record, errors.New("empty dream")
	}
	record.Dejavu = CleanResponse(fields[2])
	if record.Dejavu == "" {
		return record, errors.New("empty dejavu")
	}
	record.Conflict = CleanResponse(fields[3])
	if record.Conflict == "" {
		return record, errors.New("empty conflict")
	}
	record.Award = CleanResponse(fields[4])
	if record.Award == "" {
		return record, errors.New("empty award")
	}
	return record, nil
}

// ReadEssayRecords reads a CSV file and returns a slice of EssayRecords.
// This is best-effort; errors are logged and broken records are ignored.
func ReadEssayRecords(path string) ([]EssayRecord, error) {
	return ReadCSVRecords(path, NewEssayRecord)
}

// RandomEssayRecord returns a random EssayRecord.
func RandomEssayRecord() (EssayRecord, error) {
	records, err := ReadEssayRecords("data/original/essays.csv")
	if err != nil {
		return EssayRecord{}, fmt.Errorf("random essay record: %w", err)
	}
	return records[rand.Intn(len(records))], nil
}
