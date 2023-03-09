package data

import (
	"content-coding-gpt/pkg/openai"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

// templateCache is a cache of file paths to their corresponding templates.
var templateCache = map[string]string{}

// ConflictEssayPrompt was the writing prompt for the conflict/angry essay.
var ConflictEssayPrompt = "Imagine someone is angry with you. Why are they angry with you? What led them to be angry with you? How do you feel about the situation?"

// AwardEssayPrompt was the writing prompt for the award essay.
var AwardEssayPrompt = "Imagine that you have just received an award. What did you receive the award in? How were you able to achieve what brought you the award? How do you feel about getting it?"

// DreamEssayPrompt was the writing prompt for the dream essay.
var DreamEssayPrompt = "Imagine that you have a dream that your far-away loved one (e.g. grandmother, grandfather, parent, close friend, etc.) unexpectedly visits to say they love you and to impart life wisdom. You wake up to learn that they died the previous night. Please tell us how and why you think this happens."

// DejavuEssayPrompt was the writing prompt for the deja vu essay.
var DejavuEssayPrompt = "Imagine that you meet someone for the first time and share an uncanny sense that you've known each other for decades. Please tell us how and why you think this happened."

// EssayPrompts is a map of essay types to their writing prompts.
var EssayPrompts = map[string]string{
	"conflict": ConflictEssayPrompt,
	"angry":    ConflictEssayPrompt,
	"award":    AwardEssayPrompt,
	"dream":    DreamEssayPrompt,
	"dejavu":   DejavuEssayPrompt,
}

// Hallmarks is a map of essay types to their corresponding hallmarks.
var Hallmarks = map[string]string{
	"conflict": HumilityHallmarks,
	"angry":    HumilityHallmarks,
	"award":    HumilityHallmarks,
	"dream":    SpiritualityHallmarks,
	"dejavu":   SpiritualityHallmarks,
}

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

// ChatRequest converts an EssayRecord into an OpenAI ChatRequest.
func (r EssayRecord) ChatRequest(essayType string, temperature float32, maxTokens int) openai.ChatRequest {
	prompt := Hallmarks[essayType]
	prompt += "\nA research study participant was given the following writing prompt:\n“"
	prompt += EssayPrompts[essayType]
	prompt += "”\n\nThe participant wrote the following:\n“"
	prompt += r.SelectEssay(essayType)
	prompt += "”\n\nPlease content-code the participant's response, assessing the degree to which the "
	prompt += "participant's response is consistent with the above hallmarks. Your assessment should "
	prompt += "result in a single composite number ranging from -1.0 to 1.0, where -1.0 indicates that "
	prompt += "the participant's response is completely inconsistent with the hallmarks, 0.0 indicates "
	prompt += "that the participant's response is completely neutral with respect to the hallmarks, and "
	prompt += "1.0 indicates that the participant's response is completely consistent with the hallmarks. "
	prompt += "The composite score should be provided first, followed by reasons for your assessment.\n\n"
	return openai.ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openai.Message{
			{
				Role: openai.SYSTEM,
				Content: "You are a psychology research assistant who is content-coding " +
					"text written by participants in a research study.",
			},
			{
				Role:    openai.USER,
				Content: prompt,
			},
		},
		Temperature: temperature,
		MaxTokens:   maxTokens,
		User:        strconv.Itoa(r.ID),
	}
}

// ChatRequestTemplate converts an EssayRecord into an OpenAI ChatRequest using a
// specified template file. Use this method to experiment with different prompts.
// The templates are cached, so it's efficient to use with multiple records.
//
// Use the following optional template variables:
// - {{prompt}}: the essay prompt
// - {{essay}}: the essay response text
func (r EssayRecord) ChatRequestTemplate(essayType string, temperature float32, maxTokens int, templateFile string) (openai.ChatRequest, error) {
	// Check the cache for the template.
	template, ok := templateCache[templateFile]
	if !ok {
		// Cache miss. Read and cache the template file.
		f, err := os.Open(templateFile)
		if err != nil {
			return openai.ChatRequest{}, fmt.Errorf("error opening template file %s: %w", templateFile, err)
		}
		defer f.Close()
		b, err := io.ReadAll(f)
		if err != nil {
			return openai.ChatRequest{}, fmt.Errorf("error reading template file %s: %w", templateFile, err)
		}
		template = string(b)
		templateCache[templateFile] = template
	}

	// Replace the template variables.
	prompt := strings.ReplaceAll(template, "{{prompt}}", EssayPrompts[essayType])
	prompt = strings.ReplaceAll(prompt, "{{essay}}", r.SelectEssay(essayType))

	// Create the ChatRequest.
	return openai.ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openai.Message{
			{
				Role: openai.SYSTEM,
				Content: "You are a psychology research assistant who is content-coding " +
					"text written by participants in a research study.",
			},
			{
				Role:    openai.USER,
				Content: prompt,
			},
		},
		Temperature: temperature,
		MaxTokens:   maxTokens,
		User:        strconv.Itoa(r.ID),
	}, nil
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

// ReadEssayRecord returns the specified EssayRecord.
func ReadEssayRecord(id int) (EssayRecord, error) {
	records, err := ReadEssayRecords("data/original/essays.csv")
	if err != nil {
		return EssayRecord{}, fmt.Errorf("read essay record: %w", err)
	}
	for _, record := range records {
		if record.ID == id {
			return record, nil
		}
	}
	return EssayRecord{}, fmt.Errorf("essay record %d not found", id)
}

// RandomEssayRecord returns a random EssayRecord.
func RandomEssayRecord() (EssayRecord, error) {
	records, err := ReadEssayRecords("data/original/essays.csv")
	if err != nil {
		return EssayRecord{}, fmt.Errorf("random essay record: %w", err)
	}
	return records[rand.Intn(len(records))], nil
}

// EssayCompletion provides a request, response, and score for a single essay.
type EssayCompletion struct {
	Request  openai.ChatRequest  `json:"request"`
	Response openai.ChatResponse `json:"response"`
	Score    EssayScore          `json:"score"`
	ErrMsg   string              `json:"error,omitempty"`
}

// EssayScore contains the content-coded score for a single essay.
// pid,essay_type,essay,score,comments,duration
type EssayScore struct {
	ID        int     `csv:"pid" json:"pid"`
	EssayType string  `csv:"essay_type" json:"essay_type"`
	Essay     string  `csv:"essay" json:"essay"`
	Score     float32 `csv:"score" json:"score"`
	Comments  string  `csv:"comments" json:"comments"`
	Millis    int64   `csv:"millis" json:"millis"`
}

// CSVHeader returns the CSV header for an EssayScore.
func (s EssayScore) CSVHeader() []string {
	return []string{"pid", "essay_type", "essay", "score", "comments", "millis"}
}

// CSVFields returns the CSV fields for an EssayScore.
func (s EssayScore) CSVFields() []string {
	return []string{
		strconv.Itoa(s.ID),
		s.EssayType,
		s.Essay,
		strconv.FormatFloat(float64(s.Score), 'f', 2, 32),
		s.Comments,
		strconv.FormatInt(s.Millis, 10),
	}
}

// NewEssayScore creates a new EssayScore from an essay, essay type, chat, and duration.
func NewEssayScore(essay EssayRecord, essayType string, chat openai.ChatResponse, reverse bool, millis int64) (EssayScore, error) {
	score, err := chat.ExtractScore(reverse)
	return EssayScore{
		ID:        essay.ID,
		EssayType: essayType,
		Essay:     essay.SelectEssay(essayType),
		Score:     score,
		Comments:  chat.Choices[0].Message.Content,
		Millis:    millis,
	}, err
}

// WriteEssayScores writes a slice of EssayScores to a CSV file.
func WriteEssayScores(path string, scores []EssayScore) error {
	csvRecords := make([][]string, len(scores)+1)
	csvRecords[0] = EssayScore{}.CSVHeader()
	for i, score := range scores {
		csvRecords[i+1] = score.CSVFields()
	}
	return WriteCSVFile(path, csvRecords)
}
