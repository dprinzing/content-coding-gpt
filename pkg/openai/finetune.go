package openai

// TrainingRecord provides a prompt and an expected completion.
type TrainingRecord struct {
	// Prompt is the prompt text.
	Prompt string `json:"prompt"`

	// Completion is the completion text.
	Completion string `json:"completion"`
}

// FineTuneRequest is a request to fine-tune a model.
type FineTuneRequest struct {
	// TrainingFileID is the ID of an uploaded file containing the training data.
	TrainingFileID string `json:"training_file"`

	// ValidationFileID is the ID of an uploaded file containing the validation data.
	ValidationFileID string `json:"validation_file,omitempty"`

	// Model is the name/ID of the base model to fine-tune. The default is "curie".
	// You can select one of "ada", "babbage", "curie", "davinci", or a fine-tuned model ID.
	Model string `json:"model,omitempty"`

	// EpochCount is the number of epochs to train for. The default is 4.
	// An epoch refers to one full cycle through the training dataset.
	EpochCount int `json:"n_epochs,omitempty"`

	// BatchSize is the number of training examples to process in parallel.
	// By default, the batch size will be dynamically configured to be ~0.2%
	// of the number of examples in the training set, capped at 256.
	BatchSize int `json:"batch_size,omitempty"`

	// LearningRate is the learning rate multiplier for the fine-tuning. The fine-tuning learning
	// rate is the original learning rate used for pretraining, multiplied by this value. By
	// default, the learning rate multiplier is the 0.05, 0.1, or 0.2 depending on final batch_size.
	LearningRate float64 `json:"learning_rate_multiplier,omitempty"`

	// PromptLossWeight is the weight of the prompt loss. The default is 0.01. The weight to
	// use for loss on the prompt tokens. This controls how much the model tries to learn to
	// generate the prompt (as compared to the completion which always has a weight of 1.0),
	// and can add a stabilizing effect to training when completions are short.
	PromptLossWeight float64 `json:"prompt_loss_weight,omitempty"`

	// ComputeClassificationMetrics is a flag indicating whether to compute classification metrics.
	// The default is false. If true, the fine-tuning will compute classification metrics on the
	// validation set. This can be useful for determining whether the model is overfitting.
	ComputeClassificationMetrics bool `json:"compute_classification_metrics,omitempty"`

	// ClassificationNClasses is the number of classes to use in a classification task.
	// This parameter is required for multiclass classification tasks.
	ClassificationNClasses int `json:"classification_n_classes,omitempty"`

	// ClassificationPositiveClass is the positive class label for a binary classification task.
	// This parameter is needed to generate precision, recall, and F1 metrics when doing binary
	// classification.
	ClassificationPositiveClass string `json:"classification_positive_class,omitempty"`

	// ClassificationBetas is a list of beta values to use for computing F-beta scores. The
	// F-beta score is a generalization of F-1 score. This is only used for binary classification.
	ClassificationBetas []float64 `json:"classification_betas,omitempty"`

	// Suffix is a string of up to 40 characters that will be added to your fine-tuned model name.
	// This can be useful for distinguishing between different fine-tuned models.
	Suffix string `json:"suffix,omitempty"`
}

// FineTune provides information about an OpenAPI fine-tune job/model.
type FineTune struct {
	// ID is the fine-tune ID, e.g. "ft-AF1WoRqd3aJAHsqc9NY7iL8F".
	ID string `json:"id"`

	// Object is the object type, e.g. "fine-tune".
	Object string `json:"object"`

	// Model is the model ID, e.g. "curie".
	Model string `json:"model"`

	// FineTunedModel is the ID of the fine-tuned model.
	FineTunedModel string `json:"fine_tuned_model,omitempty"`

	// TrainingFiles is a list files containing the training data.
	TrainingFiles []File `json:"training_files"`

	// ValidationFiles is a list files containing the validation data.
	ValidationFiles []File `json:"validation_files,omitempty"`

	// ResultFiles is a list of files containing the fine-tuning results.
	ResultFiles []File `json:"result_files,omitempty"`

	// HyperParams provides hyperparameters used for fine-tuning.
	HyperParams HyperParameters `json:"hyperparams"`

	// OrganizationID is the ID of the organization that owns the fine-tune.
	OrganizationID string `json:"organization_id,omitempty"`

	// CreatedAt is a creation timestamp in epoch seconds, e.g. 1669599635.
	CreatedAt int64 `json:"created_at"`

	// UpdatedAt is an update timestamp in epoch seconds, e.g. 1669599635.
	UpdatedAt int64 `json:"updated_at"`

	// Status is the status of the fine-tune, e.g. "pending".
	Status string `json:"status"`

	// StatusDetails provides additional information about the status.
	StatusDetails interface{} `json:"status_details,omitempty"`

	// Events is a list of events that occurred during fine-tuning.
	Events []Event `json:"events,omitempty"`
}

// Name returns the fine-tune model name, or model ID if the name is not set.
func (f FineTune) Name() string {
	if f.FineTunedModel != "" {
		return f.FineTunedModel
	}
	return f.ID
}

// FineTuneList provides a list of fine-tuned models.
type FineTuneList struct {
	Object string     `json:"object"` // "list" is expected
	Data   []FineTune `json:"data"`   // list of fine-tunes
}

// Event provides information about an OpenAPI fine-tuning event.
type Event struct {
	// Object is the object type, e.g. "fine-tune-event".
	Object string `json:"object"`

	// Level is the event level, e.g. "info".
	Level string `json:"level"`

	// Message is the event message, e.g. "Job succeeded.".
	Message string `json:"message"`

	// CreatedAt is a creation timestamp in epoch seconds, e.g. 1669599635.
	CreatedAt int64 `json:"created_at"`
}

// EventList provides a list of fine-tuning events.
type EventList struct {
	Object string  `json:"object"` // "list" is expected
	Data   []Event `json:"data"`   // list of events
}

// HyperParameters provides hyperparameters for fine-tuning.
type HyperParameters struct {
	// EpochCount is the number of epochs to train for. The default is 4.
	// An epoch refers to one full cycle through the training dataset.
	EpochCount int `json:"n_epochs,omitempty"`

	// BatchSize is the number of training examples to process in parallel.
	// By default, the batch size will be dynamically configured to be ~0.2%
	// of the number of examples in the training set, capped at 256.
	BatchSize int `json:"batch_size,omitempty"`

	// PromptLossWeight is the weight of the prompt loss. The default is 0.01. The weight to
	// use for loss on the prompt tokens. This controls how much the model tries to learn to
	// generate the prompt (as compared to the completion which always has a weight of 1.0),
	// and can add a stabilizing effect to training when completions are short.
	PromptLossWeight float64 `json:"prompt_loss_weight,omitempty"`

	// LearningRate is the learning rate multiplier for the fine-tuning. The fine-tuning learning
	// rate is the original learning rate used for pretraining, multiplied by this value. By
	// default, the learning rate multiplier is the 0.05, 0.1, or 0.2 depending on final batch_size.
	LearningRate float64 `json:"learning_rate_multiplier,omitempty"`
}
