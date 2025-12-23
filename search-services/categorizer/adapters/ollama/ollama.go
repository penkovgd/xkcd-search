package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
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

	if !slices.Contains(core.GetCategories(), structOut.Category) {
		return "Other", nil
	}

	// c.log.Debug("successfully categorized", "comic", comic.URL, "categoty", structOut.Category)

	return structOut.Category, nil
}
func makePrompt(comic core.Comic) string {
	prompt := `You are a content categorization assistant for xkcd-style comics. 
	Your task is to select the SINGLE most appropriate category from the provided list.

	**INSTRUCTIONS:**
	1. Analyze the comic's metadata carefully
	2. Choose ONE category from the allowed list below
	3. Prioritize specificity - if a comic fits a specific category, choose it over broad ones
	4. Reserve "Other" ONLY when the comic genuinely doesn't align with any other category's theme
	5. Consider both the title and keywords for context
	6. Return ONLY the category name without explanations, formatting, or additional text

	**CATEGORIES WITH DESCRIPTIONS:`

	// Build categories section
	var categories []string
	for _, cat := range core.GetCategories() {
		name := string(cat)
		desc := strings.TrimSpace(core.CategoriesDesc[name])
		if desc != "" {
			categories = append(categories, fmt.Sprintf("- %s: %s", name, desc))
		} else {
			categories = append(categories, fmt.Sprintf("- %s", name))
		}
	}

	prompt += "\n" + strings.Join(categories, "\n") + "\n\n"

	prompt += `**COMIC TO CATEGORIZE:**
	Title: "` + strings.TrimSpace(comic.Title) + `"`

	if len(comic.Words) > 0 {
		prompt += "\nKeywords: " + strings.Join(comic.Words, ", ")
	}

	prompt += `

	**REMEMBER:**
	- Output ONLY the category name (e.g., "Programming & IT")
	- Do not add any other text
	- Choose the MOST SPECIFIC applicable category
	- Use "Other" sparingly - when truly no category fits`

	return prompt
}
