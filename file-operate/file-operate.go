package fileoperate

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

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
		fullPath = filepath.Join(filepath.Dir(execPath), path)
	}

	file, err := os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}

	return &File{file}, nil
}

// font Hook
// ReadMemos reads memos from disk each time it's called
func (f *File) ReadMemos() (map[uint64]Memo, error) {
	// Seek to the beginning of file
	if _, err := f.File.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	var memos []Memo
	err := sonic.ConfigDefault.NewDecoder(f.File).Decode(&memos)
	if err != nil && err != io.EOF {
		return nil, err
	}

	memoMap := make(map[uint64]Memo)
	for _, memo := range memos {
		memoMap[memo.ID] = memo
	}

	return memoMap, nil
}

// after Hook
func (f *File) writeMemos(memos map[uint64]Memo) error {
	// Reset file content
	if err := f.File.Truncate(0); err != nil {
		return err
	}
	if _, err := f.File.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// Convert map to slice for serialization
	memosSlice := make([]Memo, 0, len(memos))
	for _, memo := range memos {
		memosSlice = append(memosSlice, memo)
	}

	// Write to file
	byt, err := sonic.MarshalIndent(memosSlice, "", "  ")
	if err != nil {
		return err
	}

	_, err = f.File.Write(byt)
	return err
}

func (f *File) AppendMemo(memo Memo) error {
	memos, err := f.ReadMemos()
	if err != nil {
		return err
	}

	// Find max ID
	maxID := uint64(0)
	for id := range memos {
		if id > maxID {
			maxID = id
		}
	}
	memo.ID = maxID + 1

	// Check for duplicates
	waitToAppend := removeWhitespace(memo.Content)
	for _, v := range memos {
		if removeWhitespace(v.Content) == waitToAppend {
			return nil // Duplicate found, silently return
		}
	}
	memos[memo.ID] = memo

	// Write back to file
	return f.writeMemos(memos)
}

func (f *File) LLMReadableMemos() (string, error) {
	memos, err := f.ReadMemos()
	if err != nil {
		return "", err
	}

	sb := strings.Builder{}
	for _, memo := range memos {
		sb.WriteString(memo.String())
	}
	return sb.String(), nil
}

func (f *File) Close() error {
	return f.File.Close()
}

// 移除字符串中的空白字符
func removeWhitespace(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}
