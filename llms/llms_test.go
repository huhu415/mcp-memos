package llms

import (
	"context"
	"fmt"
	"testing"

	_ "embed"

	fileoperate "github.com/huhu415/mcp-memos/file-operate"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSearchContent(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	ac, err := NewAnthropicClient()
	assert.NoError(t, err)

	ctx := context.Background()
	des := "Mark apikey API密钥"
	file, err := fileoperate.OpenFile("muti_block_test.md")
	assert.NoError(t, err)

	result, err := ac.SearchContent(ctx, des, file.LLMReadableMemos())
	assert.NoError(t, err)

	fmt.Println(result)
}
