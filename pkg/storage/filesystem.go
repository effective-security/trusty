package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// FilesystemConnection is an adapter for local filesystem access, basically for
// testing purposes.
type FilesystemConnection struct {
}

func dirExists(path string) bool {
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		return false
	}
	return true
}

// get the full path to path, creating directories along the way if necessary
func (conn *FilesystemConnection) getPath(path string, mkdir bool) (result string, thrown error) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				thrown = err
			} else {
				panic(r)
			}
		}
	}()
	path, err := filepath.Abs(path)
	if err != nil {
		return "", errors.WithStack(err)
	}
	children := strings.Split(path, "/")
	checkDir := func(dir string) {
		if dirExists(dir) {
			return
		}
		if !mkdir {
			panic(os.ErrNotExist)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic(err)
		}
	}
	// Handle first child specially to avoid checking whether "" (or "c:") is a directory
	ret := children[0]

	for i, child := range children[1:] {
		ret += "/" + child
		// skip checking/creating the last element, which is presumably a file
		if i+1 < len(children)-1 {
			checkDir(ret)
		}
	}
	return ret, nil
}

// GetReader opens the file at path for reading and returns the handle.
func (conn *FilesystemConnection) GetReader(ctx context.Context, path string) (result io.ReadCloser, err error) {
	file, err := conn.getPath(path, false)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	fp, err := os.Open(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return fp, nil

}

// GetWriter opens the file at path for writing and returns the handle, creating
// any intermediate directories.
func (conn *FilesystemConnection) GetWriter(ctx context.Context, path string) (result io.WriteCloser, err error) {
	file, err := conn.getPath(path, true)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	fp, err := os.Create(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return fp, nil
}

// SetContentType updates object with content type, if supported
func (conn *FilesystemConnection) SetContentType(ctx context.Context, path, contentType string) error {
	return errors.New("operation not suppurted")
}

// Delete the file at path.
func (conn *FilesystemConnection) Delete(ctx context.Context, path string) error {
	file, err := conn.getPath(path, false)
	if err != nil {
		return errors.WithStack(err)
	}
	return os.Remove(file)
}

// Close does nothing for FilesystemConnection.
func (conn *FilesystemConnection) Close() error {
	return nil
}

// Wait can be used to block on the completion of a write operation.
func (conn *FilesystemConnection) Wait() error {
	return nil
}
