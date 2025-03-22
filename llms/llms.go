package llms

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/huhu415/mcp-memos/prompt"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

type AnthropicClient struct {
	*anthropic.LLM
}

func NewAnthropicClient() (*AnthropicClient, error) {
	model := os.Getenv("ANTHROPIC_MODEL")
	if model == "" {
		model = "claude-3-7-sonnet-20250219"
	}
	token := os.Getenv("LLM_TOKEN")
	baseURL := os.Getenv("LLM_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	client, err := anthropic.New(
		anthropic.WithModel(model),
		anthropic.WithToken(token),
		anthropic.WithBaseURL(baseURL),
	)
	if err != nil {
		return nil, err
	}
	return &AnthropicClient{client}, nil
}

func (ac *AnthropicClient) CompletionDescribeContent(ctx context.Context, des, content string) (string, error) {
	promdesc := strings.ReplaceAll(prompt.Completion, "{description}", des)
	promcontent := strings.ReplaceAll(promdesc, "{content}", content)
	message := []llms.MessageContent{
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: promcontent}}},
	}

	response, err := ac.GenerateContent(ctx, message,
		llms.WithMaxTokens(1000),
	)
	if err != nil {
		return "", err
	}

	choices := response.Choices
	if len(choices) < 1 {
		return "", errors.New("empty response from model")
	}
	c1 := choices[0]
	return c1.Content, nil
}

func (ac *AnthropicClient) SearchContent(ctx context.Context, des string, allContent []byte) (string, error) {
	promdesc := strings.ReplaceAll(prompt.SearchAnswer, "{muti_block}", string(allContent))
	message := []llms.MessageContent{
		{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextContent{Text: promdesc}}},
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: des}}},
	}

	response, err := ac.GenerateContent(ctx, message,
		llms.WithMaxTokens(1000),
	)
	if err != nil {
		return "", err
	}

	choices := response.Choices
	if len(choices) < 1 {
		return "", errors.New("empty response from model")
	}
	c1 := choices[0]
	return c1.Content, nil
}
