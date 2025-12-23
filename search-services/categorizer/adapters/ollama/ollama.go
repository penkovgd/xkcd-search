package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/penkovgd/closer"
	"yadro.com/course/categorizer/core"
)

type Request struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Format any    `json:"format,omitempty"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

type classifyResponse struct {
	Category string `json:"category"`
}

type Categorizer struct {
	log     *slog.Logger
	baseURL string
	model   string
	client  *http.Client
}

func New(log *slog.Logger, baseURL, model string) *Categorizer {
	// TODO ollama healthcheck
	return &Categorizer{
		log:     log,
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}
}

func (c *Categorizer) Categorize(ctx context.Context, comic core.Comic) (string, error) {
	prompt := makePrompt(comic)

	reqBody := Request{
		Model:  c.model,
		Prompt: prompt,
		Format: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"category": map[string]any{
					"type": "string",
				},
			},
			"required": []string{"category"},
		},
		Stream: false,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}
	endpoint := c.baseURL + "/api/generate"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer closer.CloseOrLog(c.log, resp.Body)

	var oResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&oResp); err != nil {
		return "", fmt.Errorf("decode ollama response: %w", err)
	}

	var structOut classifyResponse
	text := strings.TrimSpace(oResp.Response)
	if text == "" {
		return "", fmt.Errorf("empty ollama output")
	}

	if err := json.Unmarshal([]byte(text), &structOut); err != nil {
		return "", fmt.Errorf("unmarshal structured output: %w", err)
	}

	if structOut.Category == "" {
		return "", fmt.Errorf("no category in response")
	}

	// c.log.Debug("successfully categorized", "comic", comic.URL, "categoty", structOut.Category)

	return structOut.Category, nil
}
func makePrompt(comic core.Comic) string {
	var b strings.Builder

	b.WriteString("Assign exactly one category to this comic. ")
	b.WriteString("\n\n")

	var parts []string
	for _, c := range core.GetCategories() {
		name := string(c)
		desc := strings.TrimSpace(core.CategoriesDesc[name])
		if desc != "" {
			parts = append(parts, fmt.Sprintf("%s", name))
			//parts = append(parts, fmt.Sprintf("%s (%s)", name, desc))

		} else {
			parts = append(parts, name)
		}
	}
	b.WriteString("Categories: " + strings.Join(parts, ", "))
	b.WriteString("\n\n")

	b.WriteString("Title: " + strings.TrimSpace(comic.Title) + "\n")
	if len(comic.Words) > 0 {
		b.WriteString("Words: " + strings.Join(comic.Words, " ") + "\n")
	} else {
		b.WriteString("Words: \n")
	}
	return b.String()
}
