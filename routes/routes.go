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

// Version information
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
			server.WithInstructions(`You are an assistant that helps users record and retrieve text
			- When recording, make the description of the content as detailed as possible for accurate future retrieval
			- When retrieving, it's recommended to be more specific in descriptions for more accurate text retrieval
			`),
		),
		File: file,
		llm:  llm,
	}
}

func (r *Routes) Repeat() {
	tool := mcp.NewTool("repeat",
		mcp.WithDescription("Repeat user input text"),
		mcp.WithString("text", mcp.Required(), mcp.Description("Text that needs to be repeated")),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		text, ok := request.Params.Arguments["text"].(string)
		if !ok {
			return mcp.NewToolResultError("Text must be a string"), nil
		}

		return mcp.NewToolResultText(text), nil
	}

	r.McpServer.AddTool(tool, handler)
}

func (r *Routes) SaveText() {
	tool := mcp.NewTool("storeMemo",
		mcp.WithDescription("Save important text information with tags for future retrieval"),
		mcp.WithString("description", mcp.Required(), mcp.Description("Add a descriptive label for the saved text. Include context information to make retrieval easier.")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Actual text content to be saved, such as keys, code snippets, notes, etc.")),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		purpose, ok := request.Params.Arguments["description"].(string)
		if !ok {
			return mcp.NewToolResultError("Tag must be a string"), nil
		}

		text, ok := request.Params.Arguments["content"].(string)
		if !ok {
			return mcp.NewToolResultError("Content must be a string"), nil
		}

		r.File.AppendMemo(fileoperate.Memo{
			Description: purpose,
			Content:     text,
		})

		return mcp.NewToolResultText(fmt.Sprintf("Text `%s` has been saved to %s", text, r.File.Name())), nil
	}

	r.McpServer.AddTool(tool, handler)
}

func (r *Routes) SearchRelatedText() {
	tool := mcp.NewTool("retrieveMemo",
		mcp.WithDescription("Retrieve previously saved text content based on description"),
		mcp.WithString("description", mcp.Required(), mcp.Description("Enter description or descriptions related to the content you want to find, the system will return the most matching saved content")),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		description, ok := request.Params.Arguments["description"].(string)
		if !ok {
			return mcp.NewToolResultError("Description must be a string"), nil
		}
		mcp.NewLoggingMessageNotification(mcp.LoggingLevelInfo, "get_answer", description)

		memoString, err := r.File.LLMReadableMemos()
		if err != nil {
			return mcp.NewToolResultError("readMemos err"), err
		}
		answer, err := r.llm.SearchContent(ctx, description, memoString)
		if err != nil {
			return mcp.NewToolResultError("Cannot retrieve text, error:" + err.Error()), nil
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
						Text: "Call `store_memo` tool to save the following content to my memo",
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
						Text: "Okay, I will directly call the `store_memo` tool to save the content to memo",
					},
				},
				{
					Role: mcp.RoleUser,
					Content: mcp.TextContent{
						Type: "text",
						Text: "Alright, you need to call the tool now, please start",
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
