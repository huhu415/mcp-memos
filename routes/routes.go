package routes

import (
	"context"
	"fmt"

	fileoperate "github.com/huhu415/mcp-memos/file-operate"
	"github.com/huhu415/mcp-memos/llms"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
)

// 版本信息version information
var (
	BuildDate string
	GitCommit string
	Version   string
)

type Routes struct {
	McpServer *server.MCPServer
	File      *fileoperate.File
	llm       *llms.AnthropicClient
}

func NewRoutes(filePath string) *Routes {
	file, err := fileoperate.OpenFile(filePath, false)
	if err != nil {
		logrus.Panicf("Panic!!! Failed to open file: %v", err)
	}

	llm, err := llms.NewAnthropicClient()
	if err != nil {
		logrus.Panicf("Panic!!! Failed to initialize LLM: %v", err)
	}
	return &Routes{
		McpServer: server.NewMCPServer(
			"Huhu 🚀",
			Version,
			server.WithInstructions(`你是一个可以协助用户记录文本和检索文本的助手
			- 注意每次记录时, 对于记录内容的描述要尽可能详细, 以便于以后的准确检索
			- 检索时, 对于描述, 建议更具体一些, 以便于更准确地检索到相关的文本
			`),
		),
		File: file,
		llm:  llm,
	}
}

func (r *Routes) Repeat() {
	tool := mcp.NewTool("repeat",
		mcp.WithDescription("重复用户输入的文本"),
		mcp.WithString("文本", mcp.Required(), mcp.Description("需要重复的文本")),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		text, ok := request.Params.Arguments["文本"].(string)
		if !ok {
			return mcp.NewToolResultError("文本必须是一个字符串"), nil
		}

		return mcp.NewToolResultText(text), nil
	}

	r.McpServer.AddTool(tool, handler)
}

func (r *Routes) SaveText() {
	tool := mcp.NewTool("store_memo",
		mcp.WithDescription("保存重要文本信息并添加标签，方便日后检索"),
		mcp.WithString("标签", mcp.Required(), mcp.Description("为保存的文本添加一个简短描述性标签，例如'OpenAI密钥'、'Git命令'等，用于将来快速检索")),
		mcp.WithString("内容", mcp.Required(), mcp.Description("需要保存的实际文本内容，如密钥、代码片段、笔记等")),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		purpose, ok := request.Params.Arguments["标签"].(string)
		if !ok {
			return mcp.NewToolResultError("标签必须是一个字符串"), nil
		}

		text, ok := request.Params.Arguments["内容"].(string)
		if !ok {
			return mcp.NewToolResultError("内容必须是一个字符串"), nil
		}

		r.File.AppendMemo(fileoperate.Memo{
			Description: purpose,
			Content:     text,
		})

		return mcp.NewToolResultText(fmt.Sprintf("文本`%s`已保存到%s", text, r.File.Name())), nil
	}

	r.McpServer.AddTool(tool, handler)
}

func (r *Routes) SearchRelatedText() {
	tool := mcp.NewTool("retrieve_memo",
		mcp.WithDescription("根据关键词检索之前保存的文本内容"),
		mcp.WithString("关键词", mcp.Required(), mcp.Description("输入与您想查找的内容相关的关键词或描述，系统将返回最匹配的保存内容")),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		description, ok := request.Params.Arguments["关键词"].(string)
		if !ok {
			return mcp.NewToolResultError("关键词必须是一个字符串"), nil
		}
		mcp.NewLoggingMessageNotification(mcp.LoggingLevelInfo, "get_answer", description)

		memoString, err := r.File.LLMReadableMemos()
		if err != nil {
			return mcp.NewToolResultError("readMemos err"), err
		}
		answer, err := r.llm.SearchContent(ctx, description, memoString)
		if err != nil {
			return mcp.NewToolResultError("无法检索文本, 错误:" + err.Error()), nil
		}

		memos, err := r.File.ReadMemos()
		if err != nil {
			return mcp.NewToolResultError("readMemos err"), err
		}
		answerStr := ""
		for _, blockId := range answer {
			content := memos[blockId]
			answerStr += content.String()
		}

		return mcp.NewToolResultText(answerStr), nil
	}

	r.McpServer.AddTool(tool, handler)
}

func (r *Routes) AddMemoPromt() {
	handleComplexPrompt := func(
		ctx context.Context,
		request mcp.GetPromptRequest,
	) (*mcp.GetPromptResult, error) {
		arguments := request.Params.Arguments
		memo := fileoperate.Memo{
			Description: arguments["description"],
			Content:     arguments["content"],
		}
		return &mcp.GetPromptResult{
			Description: "prompt to add a memo",
			Messages: []mcp.PromptMessage{
				{
					Role: mcp.RoleUser,
					Content: mcp.TextContent{
						Type: "text",
						Text: "调用`store_memo`工具, 把以下内容保存到我的memo中",
					},
				},
				{
					Role: mcp.RoleUser,
					Content: mcp.TextContent{
						Type: "text",
						Text: memo.String(),
					},
				},
				{
					Role: mcp.RoleAssistant,
					Content: mcp.TextContent{
						Type: "text",
						Text: "好的, 我会直接调用`store_memo`工具, 把内容保存到memo中的",
					},
				},
				{
					Role: mcp.RoleUser,
					Content: mcp.TextContent{
						Type: "text",
						Text: "好了, 你需要直接调用工具了, 请直接开始",
					},
				},
			},
		}, nil
	}

	r.McpServer.AddPrompt(mcp.NewPrompt("add_memo_prompt",
		mcp.WithPromptDescription("prompt to add a memo"),
		mcp.WithArgument("description",
			mcp.ArgumentDescription("The description of the memo, for search later"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("content",
			mcp.ArgumentDescription("The content of the memo"),
			mcp.RequiredArgument(),
		),
	), handleComplexPrompt)
}
