package openai

// CompletionRequest represents a request structure for completion API.
type CompletionRequest struct {
	// Model ID to use for completion. Example: "text-davinci-003"
	Model string `json:"model"`

	// Prompt is the text to complete.
	Prompt string `json:"prompt,omitempty"`

	// Suffix is the text that comes after a completion of inserted text (optional).
	Suffix string `json:"suffix,omitempty"`

	// MaxTokens is the maximum number of tokens to generate. The default is 16.
	MaxTokens int `json:"max_tokens,omitempty"`

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

	// LogProbs instructs the API to include the log probabilities on the
	// logprobs most likely tokens, as well the chosen tokens. For example,
	// if logprobs is 5, the API will return a list of the 5 most likely tokens.
	// The API will always return the logprob of the sampled token, so there may
	// be up to logprobs+1 elements in the response. The maximum value for
	// logprobs is 5.
	LogProbs int `json:"logprobs,omitempty"`

	// Echo instructs the API to return the prompt in addition to the completion.
	// The default is false.
	Echo bool `json:"echo,omitempty"`

	// Stop is a list of up to 4 tokens that will cause the API to stop
	// generating further tokens. The default is an empty list. The returned
	// text will not contain the stop sequence. Example: ["\n\n###\n\n"]
	Stop []string `json:"stop,omitempty"`

	// PresencePenalty is a floating point value between -2.0 and 2.0 that
	// penalizes new tokens based on whether they appear in the text so far.
	// The default is 0.0.
	PresencePenalty float32 `json:"presence_penalty,omitempty"`

	// FrequencyPenalty is a floating point value between -2.0 and 2.0 that
	// penalizes new tokens based on their existing frequency in the text so
	// far. The default is 0.0.
	FrequencyPenalty float32 `json:"frequency_penalty,omitempty"`

	// BestOf is the number of different completions to request. The API will
	// return the best completion from this set. The default is 1.
	BestOf int `json:"best_of,omitempty"`

	// LogitBias is a dictionary of token to bias. Each token is associated
	// with an associated bias value ranging from -100 to 100 that biases the
	// log probabilities of that token. The default is an empty dictionary.
	LogitBias map[string]int `json:"logit_bias,omitempty"`

	// User is a unique identifier representing your end-user, which can help
	// OpenAI to monitor and detect abuse. The default is an empty string.
	User string `json:"user,omitempty"`
}

// Completion provides a predicted text completion in response to a provided
// prompt and other parameters.
type Completion struct {
	ID      string   `json:"id"`      // eg. "cmpl-6qU1OV5U2jx80TynLg6L8dmGC5kVJ"
	Object  string   `json:"object"`  // eg. "text_completion"
	Created int64    `json:"created"` // epoch seconds, eg. 1677966478
	Model   string   `json:"model"`   // eg. "text-davinci-003"
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice provides a possible text completion generated in response to a prompt.
type Choice struct {
	Text         string        `json:"text"`
	Index        int           `json:"index"`
	LogProbs     LogProbResult `json:"logprobs,omitempty"`
	FinishReason string        `json:"finish_reason"` // e.g. "length"
}

// LogProbResult provides the log probabilities of a particular Choice result.
type LogProbResult struct {
	Tokens        []string             `json:"tokens,omitempty"`
	TokenLogProbs []float32            `json:"token_logprobs,omitempty"`
	TopLogProbs   []map[string]float32 `json:"top_logprobs,omitempty"`
	TextOffset    []int                `json:"text_offset,omitempty"`
}

// Usage provides the total token usage per request to OpenAI.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}
