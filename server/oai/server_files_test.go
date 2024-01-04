package oai

import (
	"io"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

func TestFiles(t *testing.T) {
	c := newTestContext()

	file, err := c.Client.CreateFileBytes(c.Context, openai.FileBytesRequest{
		Name: "Test.txt",

		Bytes:   []byte("This is a test file"),
		Purpose: openai.PurposeAssistants,
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, file.ID)

	file, err = c.Client.GetFile(c.Context, file.ID)

	assert.NoError(t, err)
	assert.NotEmpty(t, file.ID)

	content, err := c.Client.GetFileContent(c.Context, file.ID)

	assert.NoError(t, err)

	if content != nil {
		defer content.Close()

		data, err := io.ReadAll(content)
		assert.NoError(t, err)

		assert.NotEmpty(t, data)
	}

	list, err := c.Client.ListFiles(c.Context)

	assert.NoError(t, err)
	assert.NotEmpty(t, list.Files)

	err = c.Client.DeleteFile(c.Context, file.ID)

	assert.NoError(t, err)

}
