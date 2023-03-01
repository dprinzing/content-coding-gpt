package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"sort"
	"time"
)

// Client is the OpenAI API client.
type Client struct {
	OrgID   string
	APIKey  string
	BaseURL string
	client  *http.Client
}

// NewClient instantiates a new OpenAI API client. If either orgID or apiKey
// are not provided, the environment variables OPENAI_ORG_ID and OPENAI_API_KEY
// will be used, respectively.
func NewClient(orgID, apiKey string) *Client {
	if orgID == "" {
		orgID = os.Getenv("OPENAI_ORG_ID")
	}
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	return &Client{
		OrgID:   orgID,
		APIKey:  apiKey,
		BaseURL: "https://api.openai.com/v1",
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// getRequest creates a new HTTP request with the required headers.
func (c *Client) getRequest(ctx context.Context, path string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating %s request: %w", path, err)
	}
	req.Header.Add("Accept", "application/json")
	if c.APIKey != "" {
		req.Header.Add("Authorization", "Bearer "+c.APIKey)
	}
	if c.OrgID != "" {
		req.Header.Add("OpenAI-Organization", c.OrgID)
	}
	return req, nil
}

// postRequest creates a new HTTP request with the required headers.
func (c *Client) postRequest(ctx context.Context, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("error creating %s request: %w", path, err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	if c.APIKey != "" {
		req.Header.Add("Authorization", "Bearer "+c.APIKey)
	}
	if c.OrgID != "" {
		req.Header.Add("OpenAI-Organization", c.OrgID)
	}
	return req, nil
}

// sendRequest sends the provided HTTP request and returns the response body.
func (c *Client) sendRequest(req *http.Request) ([]byte, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending %s request: %w", req.URL.Path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("%s: %s", resp.Status, req.URL.Path)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("not found: %s", req.URL.Path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error sending %s request: status code %d", req.URL.Path, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading %s response body: %w", req.URL.Path, err)
	}
	return body, nil
}

// ListModelsRaw lists the currently available models, and provides basic information
// about each one such as the owner and availability. It returns the raw JSON response.
func (c *Client) ListModelsRaw(ctx context.Context) ([]byte, error) {
	req, err := c.getRequest(ctx, "/models")
	if err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}
	return body, nil
}

// ListModels lists the currently available models, and provides basic information
// about each one such as the owner and availability.
func (c *Client) ListModels(ctx context.Context) ([]Model, error) {
	// Fetch the raw JSON response:
	body, err := c.ListModelsRaw(ctx)
	if err != nil {
		return nil, err
	}
	// Unmarshal the JSON response into a list of models:
	var list ModelList
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("list models: error unmarshaling response: %w", err)
	}
	models := list.Data
	// Sort the models by ID and return:
	sort.Slice(models, func(i, j int) bool { return models[i].ID < models[j].ID })
	return models, nil
}

// ReadModelRaw reads the details of the specified model. It returns the raw JSON response.
func (c *Client) ReadModelRaw(ctx context.Context, id string) ([]byte, error) {
	req, err := c.getRequest(ctx, "/models/"+id)
	if err != nil {
		return nil, fmt.Errorf("read model %s: %w", id, err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("read model %s: %w", id, err)
	}
	return body, nil
}

// ReadModel reads the details of the specified model.
func (c *Client) ReadModel(ctx context.Context, id string) (Model, error) {
	var model Model
	body, err := c.ReadModelRaw(ctx, id)
	if err != nil {
		return model, err
	}
	if err := json.Unmarshal(body, &model); err != nil {
		return model, fmt.Errorf("read model %s: error unmarshaling response: %w", id, err)
	}
	return model, nil
}

// UploadFile uploads a jsonl file for use with subsequent fine-tuning requests.
func (c *Client) UploadFile(ctx context.Context, fileName, purpose string, data []byte) (File, error) {
	var file File

	// Create the multipart writer
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	// File Purpose: usually "fine-tune"
	if purpose == "" {
		purpose = "fine-tune"
	}
	err := w.WriteField("purpose", purpose)
	if err != nil {
		return file, fmt.Errorf("upload file: field purpose: %w", err)
	}

	// File Name and Data
	var fw io.Writer
	fw, err = w.CreateFormFile("file", fileName)
	if err != nil {
		return file, fmt.Errorf("upload file: field file: %w", err)
	}
	_, err = io.Copy(fw, bytes.NewReader(data))
	if err != nil {
		return file, fmt.Errorf("upload file: field file: %w", err)
	}
	w.Close()

	// Create the request
	req, err := c.postRequest(ctx, "/files", &buf)
	if err != nil {
		return file, fmt.Errorf("upload file: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Send the request
	body, err := c.sendRequest(req)
	if err != nil {
		return file, fmt.Errorf("upload file: send request: %w", err)
	}
	if err := json.Unmarshal(body, &file); err != nil {
		return file, fmt.Errorf("upload file: unmarshaling response: %w", err)
	}
	return file, nil
}

// ListFilesRaw lists the currently available files, and provides basic information
// about each one such as the owner and availability. It returns the raw JSON response.
func (c *Client) ListFilesRaw(ctx context.Context) ([]byte, error) {
	req, err := c.getRequest(ctx, "/files")
	if err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}
	return body, nil
}

// ListFiles lists the currently available files, and provides basic information
// about each one such as the owner and availability.
func (c *Client) ListFiles(ctx context.Context) ([]File, error) {
	// Fetch the raw JSON response:
	body, err := c.ListFilesRaw(ctx)
	if err != nil {
		return nil, err
	}
	// Unmarshal the JSON response into a list of files:
	var list FileList
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("list files: error unmarshaling response: %w", err)
	}
	files := list.Data
	// Sort the files by name and return:
	sort.Slice(files, func(i, j int) bool { return files[i].FileName < files[j].FileName })
	return files, nil
}

// ReadFileRaw reads the metatdata detail of the specified file. It returns the raw JSON response.
func (c *Client) ReadFileRaw(ctx context.Context, id string) ([]byte, error) {
	req, err := c.getRequest(ctx, "/files/"+id)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", id, err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", id, err)
	}
	return body, nil
}

// ReadFile reads the metadata detail of the specified file.
func (c *Client) ReadFile(ctx context.Context, id string) (File, error) {
	var file File
	body, err := c.ReadFileRaw(ctx, id)
	if err != nil {
		return file, err
	}
	if err := json.Unmarshal(body, &file); err != nil {
		return file, fmt.Errorf("read file %s: error unmarshaling response: %w", id, err)
	}
	return file, nil
}

// DownloadFile reads the contents of the specified file.
func (c *Client) DownloadFile(ctx context.Context, id string) ([]byte, error) {
	req, err := c.getRequest(ctx, "/files/"+id+"/content")
	if err != nil {
		return nil, fmt.Errorf("download file %s: %w", id, err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("download file %s: %w", id, err)
	}
	return body, nil
}

// DeleteFile deletes the specified file.
func (c *Client) DeleteFile(ctx context.Context, id string) error {
	req, err := c.getRequest(ctx, "/files/"+id)
	if err != nil {
		return fmt.Errorf("delete file %s: %w", id, err)
	}
	req.Method = http.MethodDelete
	_, err = c.sendRequest(req)
	if err != nil {
		return fmt.Errorf("delete file %s: %w", id, err)
	}
	return nil
}

// CreateFineTuneRaw creates a new fine-tuned model. It returns the raw JSON response.
func (c *Client) CreateFineTuneRaw(ctx context.Context, req FineTuneRequest) ([]byte, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("create fine-tune: %w", err)
	}
	httpReq, err := c.postRequest(ctx, "/fine-tunes", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create fine-tune: %w", err)
	}
	raw, err := c.sendRequest(httpReq)
	if err != nil {
		return nil, fmt.Errorf("create fine-tune: %w", err)
	}
	return raw, nil
}

// CreateFineTune creates a new fine-tuned model. At a minimum, we should provide
// the base model ID, the training file ID, and a suffix for the new model name.
func (c *Client) CreateFineTune(ctx context.Context, req FineTuneRequest) (FineTune, error) {
	var fineTune FineTune
	body, err := c.CreateFineTuneRaw(ctx, req)
	if err != nil {
		return fineTune, err
	}
	if err := json.Unmarshal(body, &fineTune); err != nil {
		return fineTune, fmt.Errorf("create fine-tune: error unmarshaling response: %w", err)
	}
	return fineTune, nil
}

// ListFineTunesRaw lists the currently available fine-tuning jobs, and provides basic information
// about each one, including job status events. It returns the raw JSON response.
func (c *Client) ListFineTunesRaw(ctx context.Context) ([]byte, error) {
	req, err := c.getRequest(ctx, "/fine-tunes")
	if err != nil {
		return nil, fmt.Errorf("list fine-tunes: %w", err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("list fine-tunes: %w", err)
	}
	return body, nil
}

// ListFineTunes lists the currently available fine-tuning jobs, and provides basic information
// about each one, including job status events.
func (c *Client) ListFineTunes(ctx context.Context) ([]FineTune, error) {
	// Fetch the raw JSON response:
	body, err := c.ListFineTunesRaw(ctx)
	if err != nil {
		return nil, err
	}
	// Unmarshal the JSON response into a list of fine-tunes:
	var list FineTuneList
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("list fine-tunes: error unmarshaling response: %w", err)
	}
	fineTunes := list.Data
	// Sort the fine-tunes by name or ID and return:
	sort.Slice(fineTunes, func(i, j int) bool { return fineTunes[i].Name() < fineTunes[j].Name() })
	return fineTunes, nil
}

// ReadFineTuneRaw reads the metatdata detail of the specified fine-tuning job. It returns the raw JSON response.
func (c *Client) ReadFineTuneRaw(ctx context.Context, id string) ([]byte, error) {
	req, err := c.getRequest(ctx, "/fine-tunes/"+id)
	if err != nil {
		return nil, fmt.Errorf("read fine-tune %s: %w", id, err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("read fine-tune %s: %w", id, err)
	}
	return body, nil
}

// ReadFineTune reads the metadata detail of the specified fine-tuning job.
func (c *Client) ReadFineTune(ctx context.Context, id string) (FineTune, error) {
	var fineTune FineTune
	body, err := c.ReadFineTuneRaw(ctx, id)
	if err != nil {
		return fineTune, err
	}
	if err := json.Unmarshal(body, &fineTune); err != nil {
		return fineTune, fmt.Errorf("read fine-tune %s: error unmarshaling response: %w", id, err)
	}
	return fineTune, nil
}

// ListFineTuneEventsRaw lists the events for the specified fine-tuning job. It returns the raw JSON response.
func (c *Client) ListFineTuneEventsRaw(ctx context.Context, id string) ([]byte, error) {
	req, err := c.getRequest(ctx, "/fine-tunes/"+id+"/events")
	if err != nil {
		return nil, fmt.Errorf("list fine-tune events %s: %w", id, err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("list fine-tune events %s: %w", id, err)
	}
	return body, nil
}

// ListFineTuneEvents lists the events for the specified fine-tuning job.
func (c *Client) ListFineTuneEvents(ctx context.Context, id string) ([]Event, error) {
	// Fetch the raw JSON response:
	body, err := c.ListFineTuneEventsRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	// Unmarshal the JSON response into a list of fine-tune events:
	var list EventList
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("list fine-tune events %s: error unmarshaling response: %w", id, err)
	}
	return list.Data, nil
}

// CancelFineTune cancels the specified fine-tuning job.
func (c *Client) CancelFineTune(ctx context.Context, id string) (FineTune, error) {
	var fineTune FineTune
	req, err := c.getRequest(ctx, "/fine-tunes/"+id+"/cancel")
	if err != nil {
		return fineTune, fmt.Errorf("cancel fine-tune %s: %w", id, err)
	}
	req.Method = http.MethodPost
	body, err := c.sendRequest(req)
	if err != nil {
		return fineTune, fmt.Errorf("cancel fine-tune %s: %w", id, err)
	}
	if err := json.Unmarshal(body, &fineTune); err != nil {
		return fineTune, fmt.Errorf("cancel fine-tune %s: error unmarshaling response: %w", id, err)
	}
	return fineTune, nil
}

// DeleteFineTune deletes the specified fine-tuned model.
// The ID is the field FineTunedModel in the FineTune struct.
func (c *Client) DeleteFineTune(ctx context.Context, id string) error {
	req, err := c.getRequest(ctx, "/models/"+id)
	if err != nil {
		return fmt.Errorf("delete fine-tune %s: %w", id, err)
	}
	req.Method = http.MethodDelete
	_, err = c.sendRequest(req)
	if err != nil {
		return fmt.Errorf("delete fine-tune %s: %w", id, err)
	}
	return nil
}
