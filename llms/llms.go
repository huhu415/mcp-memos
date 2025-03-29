package llms

import (
	"context"
	"errors"
	"os"
	"regexp"
	"strings"

	"github.com/huhu415/mcp-memos/prompt"
	"github.com/sirupsen/logrus"
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

	logrus.Debugf("model: %s, token: %s, baseURL: %s", model, token, baseURL)

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

// 描述和内容的拓展
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

// 找到最符合的块的id
func (ac *AnthropicClient) SearchContent(ctx context.Context, des string, allContent string) (string, error) {
	promdesc := strings.ReplaceAll(prompt.SearchAnswer, "{muti_block}", allContent)
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
	c1Content := choices[0].Content

	logrus.Debugln("LLM resp: ", c1Content)

	return ac.ExtractNumberFromContent(c1Content)
}

// ExtractNumberFromContent 使用正则表达式从内容中提取数字
func (ac *AnthropicClient) ExtractNumberFromContent(content string) (string, error) {
	// 匹配一个或多个数字（包括小数点）
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(content, -1)

	if len(matches) == 0 {
		return "", errors.New("未找到任何数字")
	}

	return matches[0], nil
}
