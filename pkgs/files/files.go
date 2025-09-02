package files

import (
	"context"
	"os"
	"path/filepath"
)

type LocalFile struct {
	ctx context.Context
}

func NewLocalFileClient(ctx context.Context) (*LocalFile, error) {
	return &LocalFile{
		ctx: ctx,
	}, nil
}

func (s *LocalFile) GetValue(filePath string) (content []byte, err error) {
	_, err = os.Stat(filePath)
	if err != nil {
		return
	}
	content, err = os.ReadFile(filePath)
	return
}

func (s *LocalFile) AddVersion(filePath string, content []byte) (fileFullPath string, err error) {
	if fileFullPath, err = filepath.Abs(filePath); err != nil {
		return
	}
	fileout, err := os.Create(filePath)
	if err != nil {
		return
	}
	defer func() {
		if tmpErr := fileout.Close(); tmpErr != nil && err == nil {
			err = tmpErr
		}
	}()
	_, err = fileout.Write(content)

	return
}
