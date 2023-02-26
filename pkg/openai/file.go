package openai

// File provides information about an OpenAPI file.
type File struct {
	// ID is the file ID, e.g. "file-XjGxS3KTG0uNmNOK362iJua3".
	ID string `json:"id"`

	// Object is the object type, e.g. "file".
	Object string `json:"object"`

	// Purpose indicates the intended use of the file, e.g. "fine-tune".
	// If a purpose is supplied, the file format will be validated.
	Purpose string `json:"purpose"`

	// FileName is the name of the uploaded file, e.g. "training_data.jsonl".
	FileName string `json:"filename"`

	// Bytes is the file size in bytes, e.g. 12345.
	Bytes int `json:"bytes"`

	// CreatedAt is a creation timestamp in epoch seconds, e.g. 1669599635.
	CreatedAt int64 `json:"created_at"`

	// Status is the status of the file, e.g. "uploaded" or "processed".
	Status string `json:"status"`

	// StatusDetails provides additional information about the status.
	StatusDetails interface{} `json:"status_details,omitempty"`
}

// FileList is a list of files that belong to the user's organization.
type FileList struct {
	Object string `json:"object"` // "list" is expected
	Data   []File `json:"data"`   // list of files
}
