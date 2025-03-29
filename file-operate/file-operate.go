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

func (m *Memo) String() string {
	return fmt.Sprintf("id: `%d`\n描述: `%s`\n内容: ```\n%s\n```\n\n ------------ \n\n", m.ID, m.Description, m.Content)
}

type File struct {
	*os.File
	memos map[uint64]Memo
}

func OpenFile(path string, fullPathOK bool) (*File, error) {
	var fullPath string
	if fullPathOK {
		fullPath = path
	} else {
		execPath, err := os.Executable()
		if err != nil {
			return nil, err
		}
		fullPath = filepath.Join(execPath, path)
	}

	file, err := os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}

	var memos []Memo
	sonic.ConfigDefault.NewDecoder(file).Decode(&memos)

	memoMap := make(map[uint64]Memo)
	for _, memo := range memos {
		memoMap[memo.ID] = memo
	}

	return &File{file, memoMap}, nil
}

func (f *File) ReadMemos() map[uint64]Memo {
	return f.memos
}

func (f *File) AppendMemo(memo Memo) {
	memo.ID = f.findMaxID() + 1
	f.memos[memo.ID] = memo
}

func (f *File) LLMReadableMemos() string {
	sb := strings.Builder{}
	for _, memo := range f.memos {
		sb.WriteString(memo.String())
	}
	return sb.String()
}

func (f *File) findMaxID() uint64 {
	maxID := uint64(0)
	for id := range f.memos {
		if id > maxID {
			maxID = id
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

	memos := make([]Memo, 0, len(f.memos))
	for _, memo := range f.memos {
		memos = append(memos, memo)
	}

	byt, err := sonic.MarshalIndent(memos, "", "  ")
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
