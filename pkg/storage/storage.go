package storage

import (
	"context"
	"io"
	"io/ioutil"
	"strings"

	"github.com/effective-security/xlog"
	"github.com/pkg/errors"
)

var logger = xlog.NewPackageLogger("github.com/effective-security/trusty/pkg", "storage")

// Options collects the options for all supported storage types
type Options struct {
	GoogleOptions
}

// ReadConnection is a connection that can return file readers.
type ReadConnection interface {
	io.Closer

	// GetReader returns a reader for path; the caller is responsible for
	// closing the reader when finished.
	GetReader(ctx context.Context, path string) (io.ReadCloser, error)

	// Wait for something to complete after close
	Wait() error
}

// WriteConnection is a connection that can return file writers.
type WriteConnection interface {
	io.Closer

	// GetWriter returns a writer for path; the caller is responsible for
	// closing the writer when finished.
	GetWriter(ctx context.Context, path string) (io.WriteCloser, error)

	// Delete the file pointed to by path.
	Delete(ctx context.Context, path string) error

	// Wait for something to complete after close
	Wait() error
}

// ReadWriteConnection is a connection that can return both file readers and
// file writers.
type ReadWriteConnection interface {
	io.Closer

	// GetReader returns a reader for path; the caller is responsible for
	// closing the reader when finished.
	GetReader(ctx context.Context, path string) (io.ReadCloser, error)

	// GetWriter returns a writer for path; the caller is responsible for
	// closing the writer when finished.
	GetWriter(ctx context.Context, path string) (io.WriteCloser, error)

	// Delete the file pointed to by path.
	Delete(ctx context.Context, path string) error

	// Wait for something to complete after close
	Wait() error
}

// ConnectionFromPath attempts to deduce the right type of storage connection
// from the given path. Currently this works because the only supported types
// are Google Cloud Storage, whose paths require a gs:// prefix, and filesystem.
func ConnectionFromPath(path string, options ...*Options) (ReadWriteConnection, error) {
	if strings.HasPrefix(path, "gs://") {
		gcs := GoogleOptions{}
		if len(options) > 0 {
			gcs = options[0].GoogleOptions
		}
		return &GoogleCloudStorageConnection{
			GoogleOptions: gcs,
		}, nil
	}
	return &FilesystemConnection{}, nil
}

// WrappedReader is an io.ReadCloser that carries its connection with it, and
// WrappedReader.Close() closes both the reader and the connection.
type WrappedReader struct {
	Conn   ReadWriteConnection
	Reader io.ReadCloser
}

// WrappedWriter is an io.WriterCloser that carries its connection with it, and
// WrappedWriter.Close() closes both the writer and the connection.
type WrappedWriter struct {
	Conn   ReadWriteConnection
	Writer io.WriteCloser
}

// Read from the underlying reader.
func (reader *WrappedReader) Read(buf []byte) (int, error) {
	return reader.Reader.Read(buf)
}

// Close the underlying reader and then close the connection. If either returns
// an error, return whichever occurred first.
func (reader *WrappedReader) Close() error {
	readerError := reader.Reader.Close()
	connError := reader.Conn.Close()
	if readerError != nil {
		return errors.WithStack(readerError)
	}
	if connError != nil {
		return errors.WithStack(connError)
	}
	return nil
}

// Write buf to the underlying connection.
func (writer *WrappedWriter) Write(buf []byte) (int, error) {
	return writer.Writer.Write(buf)
}

// Close the underlying writer and then close the connection. If either returns
// an error, return whichever occurred first.
func (writer *WrappedWriter) Close() error {
	writerError := writer.Writer.Close()
	connError := writer.Conn.Close()
	if writerError != nil {
		return errors.WithStack(writerError)
	}
	return connError
}

// GetReaderFromPath returns a reader on a singleton connection to the given
// path. Closing the reader closes the connection.
func GetReaderFromPath(ctx context.Context, path string, options ...*Options) (*WrappedReader, error) {
	conn, err := ConnectionFromPath(path, options...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return GetReaderFromConn(ctx, path, conn)
}

// GetReaderFromConn is like GetReaderFromPath, but where the caller had made an explicit connection
func GetReaderFromConn(ctx context.Context, path string, conn ReadWriteConnection) (*WrappedReader, error) {
	reader, err := conn.GetReader(ctx, path)
	if err != nil {
		conn.Close()
		return nil, errors.WithStack(err)
	}
	return &WrappedReader{
		Conn:   conn,
		Reader: reader,
	}, nil
}

// GetWriterFromPath returns a writer on a singleton connection to the given
// path. Closing the reader closes the connection.
func GetWriterFromPath(ctx context.Context, path string, options ...*Options) (*WrappedWriter, error) {
	conn, err := ConnectionFromPath(path, options...)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to create connection: %s", path)
	}
	return GetWriterFromConn(ctx, path, conn)
}

// GetWriterFromConn is like GetWriterFromPath, but where the caller had made an explicit connection
func GetWriterFromConn(ctx context.Context, path string, conn ReadWriteConnection) (*WrappedWriter, error) {
	writer, err := conn.GetWriter(ctx, path)
	if err != nil {
		conn.Close()
		return nil, errors.WithMessagef(err, "failed to create writer: %s", path)
	}
	return &WrappedWriter{
		Conn:   conn,
		Writer: writer,
	}, nil
}

// DeletePath deletes the file at path using a singleton connection that is
// immediately closed after deletion. If either Delete or Close return an error,
// the first is returned.
func DeletePath(ctx context.Context, path string, options ...*Options) error {
	conn, err := ConnectionFromPath(path, options...)
	if err != nil {
		return errors.WithStack(err)
	}
	deleteError := conn.Delete(ctx, path)
	closeError := conn.Close()
	if deleteError != nil {
		return errors.WithStack(deleteError)
	}
	return closeError
}

// ReadFile reads the entire contents of path on a singleton connection that is
// closed before returning. NOTE: Close errors are not returned.
func ReadFile(ctx context.Context, path string, options ...*Options) ([]byte, error) {
	reader, err := GetReaderFromPath(ctx, path, options...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer reader.Close()
	ret, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to read: %s", path)
	}
	return ret, nil
}

// WriteFile writes data to path on a singleton connection that is closed before
// returning, and returns the amount of data written (or the first error
// encountered while trying). If all of data is not able to be written, returns
// io.ErrShortWrite.
func WriteFile(ctx context.Context, path string, data []byte, options ...*Options) (int, error) {
	writer, err := GetWriterFromPath(ctx, path, options...)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	defer writer.Close()
	n, err := writer.Write(data)
	if err != nil {
		return n, errors.WithMessagef(err, "failed to write: %s", path)
	}
	if n != len(data) {
		err = errors.WithStack(io.ErrShortWrite)
	}
	return n, err
}
