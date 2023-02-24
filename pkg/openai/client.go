package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

// addHeaders adds the required headers to the provided HTTP request.
func (c *Client) addHeaders(req *http.Request) error {
	if req == nil {
		return fmt.Errorf("add headers: request is nil")
	}
	req.Header.Add("Accept", "application/json")
	if req.Method == http.MethodPost {
		req.Header.Add("Content-Type", "application/json")
	}
	if c.APIKey != "" {
		req.Header.Add("Authorization", "Bearer "+c.APIKey)
	}
	if c.OrgID != "" {
		req.Header.Add("OpenAI-Organization", c.OrgID)
	}
	return nil
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

// ListModelsRaw lists the currently available models, and provides basic information
// about each one such as the owner and availability. It returns the raw JSON response.
func (c *Client) ListModelsRaw(ctx context.Context) ([]byte, error) {
	// Create the request:
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("list models: error creating request: %w", err)
	}
	// Add the required headers:
	if err := c.addHeaders(req); err != nil {
		return nil, fmt.Errorf("list models: error adding headers: %w", err)
	}
	// Send the request:
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list models: error sending request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list models: unexpected status code: %s", resp.Status)
	}
	// Read the response body:
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("list models: error reading response body: %w", err)
	}
	return body, nil
}

// ReadModel reads the details of the specified model.
func (c *Client) ReadModel(ctx context.Context, modelID string) (Model, error) {
	var model Model
	// Fetch the raw JSON response:
	body, err := c.ReadModelRaw(ctx, modelID)
	if err != nil {
		return model, err
	}
	// Unmarshal the JSON response into a model:
	if err := json.Unmarshal(body, &model); err != nil {
		return model, fmt.Errorf("read model: error unmarshaling response: %w", err)
	}
	return model, nil
}

// ReadModelRaw reads the details of the specified model. It returns the raw JSON response.
func (c *Client) ReadModelRaw(ctx context.Context, modelID string) ([]byte, error) {
	// Create the request:
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/models/"+modelID, nil)
	if err != nil {
		return nil, fmt.Errorf("read model: error creating request: %w", err)
	}
	// Add the required headers:
	if err := c.addHeaders(req); err != nil {
		return nil, fmt.Errorf("read model: error adding headers: %w", err)
	}
	// Send the request:
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("read model: error sending request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("read model: model not found: %s", modelID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("read model: unexpected status code: %s", resp.Status)
	}
	// Read the response body:
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read model: error reading response body: %w", err)
	}
	return body, nil
}
