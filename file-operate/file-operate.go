package fileoperate

import (
	"io"
	"os"
	"path/filepath"
)

type File struct {
	*os.File
}

func OpenFile(path string) (*File, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(execPath)

	fullPath := filepath.Join(dir, path)

	file, err := os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &File{file}, nil
}

func (f *File) Read() ([]byte, error) {
	return io.ReadAll(f.File)
}

func (f *File) Write(data []byte) error {
	_, err := f.File.Write(data)
	return err
}
