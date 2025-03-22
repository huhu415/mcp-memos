package llms

import (
	"context"
	"fmt"
	"os"
	"testing"

	_ "embed"

	"github.com/stretchr/testify/assert"
)

func TestSearchContent(t *testing.T) {
	ac, err := NewAnthropicClient()
	assert.NoError(t, err)

	ctx := context.Background()
	des := "Mark apikey API密钥"
	content, err := os.ReadFile("muti_block_test.md")
	assert.NoError(t, err)

	result, err := ac.SearchContent(ctx, des, content)
	assert.NoError(t, err)

	fmt.Println(result)
}
