package openai

import (
	"errors"
	"strconv"
	"strings"
)

// ChatRequest represents a request structure for chat completion API.
type ChatRequest struct {
	// Model ID to use for completion. Example: "gpt-3.5-turbo"
	Model string `json:"model"`

	// Messages is a list of messages in the conversation.
	Messages []Message `json:"messages"`

	// Temperature is the sampling temperature. Higher values result in more
	// random completions. Values range between 0 and 2. Higher values like
	// 0.8 will make the output more random, while lower values like 0.2 will
	// make it more focused and deterministic. The default is 1.0.
	Temperature float32 `json:"temperature,omitempty"`

	// TopP is the top-p sampling parameter. If set to a value between 0 and 1,
	// the returned text will be sampled from the smallest possible set of
	// tokens whose cumulative probability exceeds the value of top_p. For
	// example, if top_p is set to 0.1, the API will only consider the top 10%
	// probability tokens each step. This can be used to ensure that the
	// returned text doesn't contain undesirable tokens. The default is 1.0.
	TopP float32 `json:"top_p,omitempty"`

	// N is the number of results to return. The default is 1.
	N int `json:"n,omitempty"`

	// Stream is whether to stream back partial progress. The default is false.
	Stream bool `json:"stream,omitempty"`

	// Stop is a list of up to 4 tokens that will cause the API to stop
	// generating further tokens. The default is an empty list. The returned
	// text will not contain the stop sequence. Example: ["\n\n###\n\n"]
	Stop []string `json:"stop,omitempty"`

	// MaxTokens is the maximum number of tokens to generate. The default is
	// "infinity", but the actual default maximum is (4096 - prompt tokens).
	MaxTokens int `json:"max_tokens,omitempty"`

	// PresencePenalty is a floating point value between -2.0 and 2.0 that
	// penalizes new tokens based on whether they appear in the text so far.
	// The default is 0.0.
	PresencePenalty float32 `json:"presence_penalty,omitempty"`

	// FrequencyPenalty is a floating point value between -2.0 and 2.0 that
	// penalizes new tokens based on their existing frequency in the text so
	// far. The default is 0.0.
	FrequencyPenalty float32 `json:"frequency_penalty,omitempty"`

	// LogitBias is a dictionary of token to bias. Each token is associated
	// with an associated bias value ranging from -100 to 100 that biases the
	// log probabilities of that token. The default is an empty dictionary.
	LogitBias map[string]int `json:"logit_bias,omitempty"`

	// User is a unique identifier representing your end-user, which can help
	// OpenAI to monitor and detect abuse. The default is an empty string.
	User string `json:"user,omitempty"`
}

// Chat provides a predicted text completion in response to a provided
// prompt and other parameters.
type Chat struct {
	ID      string          `json:"id"`      // eg. "chatcmpl-6p9XYPYSTTRi0xEviKjjilqrWU2Ve"
	Object  string          `json:"object"`  // eg. "chat.completion"
	Created int64           `json:"created"` // epoch seconds, eg. 1677966478
	Model   string          `json:"model"`   // eg. "gpt-3.5-turbo"
	Usage   Usage           `json:"usage"`
	Choices []MessageChoice `json:"choices"`
}

// ExtractScore returns the first floating-point score found in the first choice.
func (c *Chat) ExtractScore() (float32, error) {
	if len(c.Choices) == 0 {
		return 0, errors.New("chat score: no choices found")
	}
	if len(c.Choices[0].Message.Content) == 0 {
		return 0, errors.New("chat score: no content found")
	}
	fields := strings.Fields(c.Choices[0].Message.Content)
	if len(fields) == 0 {
		return 0, errors.New("chat score: no words found")
	}
	for _, field := range fields {
		if score, err := ParseScore(field); err == nil {
			return score, nil
		}
	}
	return 0, errors.New("chat score: no score found")
}

// ParseScore parses a string as a floating-point number.
func ParseScore(s string) (float32, error) {
	// Check if the word starts with a minus sign or numeric digit:
	if len(s) == 0 || (s[0] != '-' && (s[0] < '0' || s[0] > '9')) {
		return 0, errors.New("score: not a number")
	}
	// Remove trailing punctuation:
	for len(s) > 0 && (s[len(s)-1] < '0' || s[len(s)-1] > '9') {
		s = s[:len(s)-1]
	}
	// Parse the number:
	score, err := strconv.ParseFloat(s, 32)
	return float32(score), err
}

// MessageChoice represents a choice in a chat completion.
type MessageChoice struct {
	Message      Message `json:"message"`
	Index        int     `json:"index"`
	FinishReason string  `json:"finish_reason"` // e.g. "stop"
}

// Message represents a message in a chat conversation.
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}
