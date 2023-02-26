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
func (c *Client) ReadModelRaw(ctx context.Context, modelID string) ([]byte, error) {
	req, err := c.getRequest(ctx, "/models/"+modelID)
	if err != nil {
		return nil, fmt.Errorf("read model %s: %w", modelID, err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("read model %s: %w", modelID, err)
	}
	return body, nil
}

// ReadModel reads the details of the specified model.
func (c *Client) ReadModel(ctx context.Context, modelID string) (Model, error) {
	var model Model
	body, err := c.ReadModelRaw(ctx, modelID)
	if err != nil {
		return model, err
	}
	if err := json.Unmarshal(body, &model); err != nil {
		return model, fmt.Errorf("read model %s: error unmarshaling response: %w", modelID, err)
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
func (c *Client) ReadFileRaw(ctx context.Context, fileID string) ([]byte, error) {
	req, err := c.getRequest(ctx, "/files/"+fileID)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", fileID, err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", fileID, err)
	}
	return body, nil
}

// ReadFile reads the metadata detail of the specified file.
func (c *Client) ReadFile(ctx context.Context, fileID string) (File, error) {
	var file File
	body, err := c.ReadFileRaw(ctx, fileID)
	if err != nil {
		return file, err
	}
	if err := json.Unmarshal(body, &file); err != nil {
		return file, fmt.Errorf("read file %s: error unmarshaling response: %w", fileID, err)
	}
	return file, nil
}

// DownloadFile reads the contents of the specified file.
func (c *Client) DownloadFile(ctx context.Context, fileID string) ([]byte, error) {
	req, err := c.getRequest(ctx, "/files/"+fileID+"/content")
	if err != nil {
		return nil, fmt.Errorf("download file %s: %w", fileID, err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("download file %s: %w", fileID, err)
	}
	return body, nil
}

// DeleteFile deletes the specified file.
func (c *Client) DeleteFile(ctx context.Context, fileID string) error {
	req, err := c.getRequest(ctx, "/files/"+fileID)
	if err != nil {
		return fmt.Errorf("delete file %s: %w", fileID, err)
	}
	req.Method = http.MethodDelete
	_, err = c.sendRequest(req)
	if err != nil {
		return fmt.Errorf("delete file %s: %w", fileID, err)
	}
	return nil
}
