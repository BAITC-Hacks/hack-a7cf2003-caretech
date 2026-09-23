package llm

import (
	"os"
	"testing"
)

func TestNewClientFromEnvLoadsDotEnv(t *testing.T) {
	tmpDir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(oldWD)
	}()

	if err := os.WriteFile(".env", []byte("LLM_BASE_URL=http://localhost:11434/v1/chat/completions\nLLM_MODEL=qwen2.5:7b\nLLM_API_KEY=ollama\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, key := range []string{"LLM_API_KEY", "LLM_BASE_URL", "LLM_MODEL", "NVIDIA_API_KEY", "NVIDIA_API_BASE_URL", "NVIDIA_MODEL"} {
		_ = os.Unsetenv(key)
	}

	clientInstance := NewClientFromEnv()
	if clientInstance == nil {
		t.Fatal("expected client to be initialized from .env")
	}
	c, ok := clientInstance.(*client)
	if !ok {
		t.Fatalf("expected concrete client type, got %T", clientInstance)
	}
	if c.baseURL != "http://localhost:11434/v1/chat/completions" {
		t.Fatalf("unexpected baseURL: %q", c.baseURL)
	}
	if c.model != "qwen2.5:7b" {
		t.Fatalf("unexpected model: %q", c.model)
	}
	if c.apiKey != "ollama" {
		t.Fatalf("unexpected apiKey: %q", c.apiKey)
	}
}
