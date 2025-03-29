package llms

import (
	"context"
	"errors"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/bytedance/sonic"
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
func (ac *AnthropicClient) SearchContent(ctx context.Context, des string, allContent string) ([]uint64, error) {
	promdesc := strings.ReplaceAll(prompt.SearchAnswer, "{muti_block}", allContent)
	message := []llms.MessageContent{
		{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextContent{Text: promdesc}}},
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: des}}},
	}

	response, err := ac.GenerateContent(ctx, message,
		llms.WithMaxTokens(1000),
	)
	if err != nil {
		return nil, err
	}

	choices := response.Choices
	if len(choices) < 1 {
		return nil, errors.New("empty response from model")
	}
	c1Content := choices[0].Content
	logrus.Debugln("LLM first resp: ", c1Content)

	blockId := ac.ExtractNumberFromContent(c1Content)
	switch len(blockId) {
	case 0:
		return nil, errors.New("未找到任何数字")
	case 1:
		// 继续往下运行
	default:
		jsonContent, err := ac.jsonFormat(ctx, message, c1Content)
		if err != nil {
			return nil, err
		}
		if blockId, err = ac.findJson(jsonContent); err != nil {
			return nil, err
		}
	}

	return blockId, nil
}

func (ac *AnthropicClient) jsonFormat(ctx context.Context, message []llms.MessageContent, content string) (string, error) {
	message = append(message, []llms.MessageContent{
		{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{llms.TextContent{Text: content}}},
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: "使用json数组回答"}}},
	}...)

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

	return c1Content, nil
}

// ExtractNumberFromContent 使用正则表达式从内容中提取数字
func (ac *AnthropicClient) ExtractNumberFromContent(content string) []uint64 {
	// 匹配一个或多个数字（包括小数点）
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(content, -1)

	blockIds := make([]uint64, 0, len(matches))
	for _, match := range matches {
		blockId, err := strconv.ParseUint(match, 10, 64)
		if err != nil {
			continue
		}
		blockIds = append(blockIds, blockId)
	}
	return blockIds
}

// 在一个字符串中, 找到第一个‘[’和最后一个‘]’之间的内容
func (ac *AnthropicClient) findJson(content string) ([]uint64, error) {
	re := regexp.MustCompile(`\[(.*?)\]`)
	matches := re.FindStringSubmatch(content)
	if len(matches) != 2 {
		return nil, errors.New("there no have [ ]")
	}
	jsonContent := matches[0]

	var val []uint64
	if err := sonic.UnmarshalString(jsonContent, &val); err != nil {
		return nil, err
	}
	return val, nil
}
