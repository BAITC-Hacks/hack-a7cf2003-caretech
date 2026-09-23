package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type Client interface {
	Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type client struct {
	apiKey   string
	baseURL  string
	model    string
	httpCli  HTTPClient
}

func NewClientFromEnv() Client {
	apiKey := strings.TrimSpace(os.Getenv("NVIDIA_API_KEY"))
	if apiKey == "" {
		return nil
	}
	baseURL := strings.TrimSpace(os.Getenv("NVIDIA_API_BASE_URL"))
	if baseURL == "" {
		baseURL = "https://integrate.api.nvidia.com/v1/chat/completions"
	}
	model := strings.TrimSpace(os.Getenv("NVIDIA_MODEL"))
	if model == "" {
		model = "meta/llama-3.1-70b-instruct"
	}
	return &client{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		httpCli: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *client) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("llm client is not configured")
	}
	payload := map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.2,
		"max_tokens":  300,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		var errBody struct {
			Error map[string]any `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		if errBody.Error != nil {
			return "", fmt.Errorf("nvidia api error: %v", errBody.Error)
		}
		return "", fmt.Errorf("nvidia api error: status %d", resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("nvidia api returned no choices")
	}
	answer := strings.TrimSpace(result.Choices[0].Message.Content)
	if answer == "" {
		return "", fmt.Errorf("nvidia api returned empty answer")
	}
	return answer, nil
}
