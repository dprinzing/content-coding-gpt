package openai

type TrainingRecord struct {
	// Prompt is the prompt text.
	Prompt string `json:"prompt"`

	// Completion is the completion text.
	Completion string `json:"completion"`
}
