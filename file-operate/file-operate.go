package fileoperate

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bytedance/sonic"
)

type Memo struct {
	ID          uint64 `json:"id"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

type File struct {
	*os.File
	memos []Memo
}

func OpenFile(path string) (*File, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(execPath)
	fullPath := filepath.Join(dir, path)
	file, err := os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}

	var memos []Memo
	sonic.ConfigDefault.NewDecoder(file).Decode(&memos)

	return &File{file, memos}, nil
}

func (f *File) ReadMemos() []Memo {
	return f.memos
}

func (f *File) AppendMemo(memo Memo) {
	memo.ID = f.findMaxID() + 1
	f.memos = append(f.memos, memo)
}

func (f *File) LLMReadableMemos() string {
	sb := strings.Builder{}
	for _, memo := range f.memos {
		sb.WriteString(fmt.Sprintf("id: `%d`\n描述: `%s`\n内容: ```\n%s\n```\n\n ------------ \n\n", memo.ID, memo.Description, memo.Content))
	}
	return sb.String()
}

func (f *File) findMaxID() uint64 {
	maxID := uint64(0)
	for _, memo := range f.memos {
		if memo.ID > maxID {
			maxID = memo.ID
		}
	}
	return maxID
}

func (f *File) WriteToFile() error {
	if err := f.File.Sync(); err != nil {
		return err
	}
	if err := f.File.Truncate(0); err != nil {
		return err
	}
	if _, err := f.File.Seek(0, io.SeekStart); err != nil {
		return err
	}

	byt, err := sonic.MarshalIndent(f.memos, "", "  ")
	if err != nil {
		return err
	}

	_, err = f.File.Write(byt)
	return err
}

func (f *File) Close() error {
	if err := f.WriteToFile(); err != nil {
		return err
	}
	return f.File.Close()
}
