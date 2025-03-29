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
	logrus.SetReportCaller(true)

	ac, err := NewAnthropicClient()
	assert.NoError(t, err)

	ctx := context.Background()
	des := "Mark apikey API密钥"
	file, err := fileoperate.OpenFile("/Users/hello/projects/mcp-memos/llms/muti_block_test.json", true)
	assert.NoError(t, err)

	logrus.Debugln("file: ", file.Name())

	result, err := ac.SearchContent(ctx, des, file.LLMReadableMemos())
	assert.NoError(t, err)

	fmt.Println(result)
}
