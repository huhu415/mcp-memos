package main

import (
	"fmt"

	"github.com/huhu415/mcp-memos/routes"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	r := routes.NewRoutes("memos.json")
	defer r.File.Close()

	r.SaveText()
	r.Repeat()
	r.SearchRelatedText()

	r.AddMemoPromt()

	if err := server.ServeStdio(r.McpServer); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
