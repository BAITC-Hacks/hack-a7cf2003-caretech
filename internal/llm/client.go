package llm

import (
	"bufio"
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
	apiKey  string
	baseURL string
	model   string
	httpCli HTTPClient
}

func loadDotEnv() {
	paths := []string{".env", ".env.local"}
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			value = strings.Trim(value, "\"'")
			if _, exists := os.LookupEnv(key); !exists {
				_ = os.Setenv(key, value)
			}
		}
		_ = scanner.Err()
	}
}

func NewClientFromEnv() Client {
	loadDotEnv()

	apiKey := strings.TrimSpace(os.Getenv("LLM_API_KEY"))
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("NVIDIA_API_KEY"))
	}
	baseURL := strings.TrimSpace(os.Getenv("LLM_BASE_URL"))
	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv("NVIDIA_API_BASE_URL"))
	}
	if baseURL == "" {
		baseURL = "http://localhost:11434/v1/chat/completions"
	}
	model := strings.TrimSpace(os.Getenv("LLM_MODEL"))
	if model == "" {
		model = strings.TrimSpace(os.Getenv("NVIDIA_MODEL"))
	}
	if model == "" {
		model = "qwen2.5:7b"
	}
	if apiKey == "" {
		if strings.HasPrefix(baseURL, "http://localhost:") || strings.HasPrefix(baseURL, "http://127.0.0.1:") || strings.HasPrefix(baseURL, "http://host.docker.internal:") {
			apiKey = "ollama"
		} else {
			return nil
		}
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
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

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
