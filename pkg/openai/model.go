package openai

// Model identifies an OpenAPI model.
type Model struct {
	// ID is the model ID, e.g. "text-davinci-003".
	ID string `json:"id"`

	// Object is the object type, e.g. "model".
	Object string `json:"object"`

	// Created is a creation timestamp in epoch seconds, e.g. 1669599635.
	Created int64 `json:"created"`

	// OwnedBy is the owner of the model, e.g. "openai-internal".
	OwnedBy string `json:"owned_by"`

	// Permissions are the model permissions. There is typically one permission.
	Permissions []Permission `json:"permission"`

	// Root is the base model ID, e.g. "text-davinci-003".
	Root string `json:"root"`

	// Parent is the parent model ID, e.g. null.
	Parent string `json:"parent"`
}

// Permission indicates how and by whom an OpenAPI Model may be used.
type Permission struct {
	// ID is the permission ID, e.g. "modelperm-loLaKHUdKtFOPD6zujUCDHno".
	ID string `json:"id"`

	// Object is the object type, e.g. "model_permission".
	Object string `json:"object"`

	// Created is a creation timestamp in epoch seconds, e.g. 1677093237.
	Created int64 `json:"created"`

	// AllowCreateEngine indicates whether the model may be used to create an engine.
	AllowCreateEngine bool `json:"allow_create_engine"`

	// AllowSampling indicates whether the model may be used to sample text.
	AllowSampling bool `json:"allow_sampling"`

	// AllowLogprobs indicates whether the model may be used to return log probabilities.
	AllowLogprobs bool `json:"allow_logprobs"`

	// AllowSearchIndices indicates whether the model may be used to search indices.
	AllowSearchIndices bool `json:"allow_search_indices"`

	// AllowView indicates whether the model may be viewed.
	AllowView bool `json:"allow_view"`

	// AllowFineTuning indicates whether the model may be used to fine-tune a new model.
	AllowFineTuning bool `json:"allow_fine_tuning"`

	// Organization indicates which organization(s) may use the model. Example: "*" for everyone.
	Organization string `json:"organization"`

	// Group? Not sure. Example: null.
	Group interface{} `json:"group"`

	// IsBlocking indicates whether the model is blocking. Example: false.
	IsBlocking bool `json:"is_blocking"`
}

type ModelList struct {
	Object string  `json:"object"` // "list" is expected
	Data   []Model `json:"data"`   // list of models
}
