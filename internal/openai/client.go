package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/dave1010/jorin/internal/types"
)

// LLM is an interface representing a language model provider that can
// produce a single chat completion. This abstraction lets us add other
// LLM backends later without duplicating session orchestration logic.
type LLM interface {
	ChatOnce(model string, msgs []types.Message, toolsList []types.Tool) (*types.ChatResponse, error)
}

// DefaultLLM is the package-level LLM implementation used by the
// convenience functions in this package. It defaults to the OpenAI
// HTTP client implementation below.
var DefaultLLM LLM = openAIClient{}

type openAIClient struct{}

func openAIBase() string {
	if b := os.Getenv("OPENAI_BASE_URL"); b != "" {
		return strings.TrimRight(b, "/")
	}
	return "https://api.openai.com"
}

func (o openAIClient) ChatOnce(model string, msgs []types.Message, toolsList []types.Tool) (*types.ChatResponse, error) {
	body := types.ChatRequest{
		Model:      model,
		Messages:   msgs,
		Tools:      toolsList,
		ToolChoice: "auto",
	}
	j, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", openAIBase()+"/v1/chat/completions", bytes.NewReader(j))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			// best-effort close; nothing useful to do here
		}
	}()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API %d: %s", resp.StatusCode, string(b))
	}
	var out types.ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
