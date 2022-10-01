package storage

import (
	"context"
	"io"
	"net/url"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/effective-security/xlog"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
)

// GoogleOptions are command-line options for the connection to Google Cloud
// Storage.
type GoogleOptions struct {
	APIKey          string
	CredentialsFile string
	ConnectTimeout  time.Duration
}

// GetClientOptions returns the ClientOption slice for these options.
func (opts *GoogleOptions) GetClientOptions() []option.ClientOption {
	ret := []option.ClientOption{}
	// If none of these are specified, it should use either the
	// GOOGLE_APPLICATION_CREDENTIALS environment variable or the metadata
	// server on the Google Compute engine.
	if opts.APIKey != "" {
		ret = append(ret, option.WithAPIKey(opts.APIKey))
	}
	if opts.CredentialsFile != "" {
		ret = append(ret, option.WithCredentialsFile(opts.CredentialsFile))
	}
	return ret
}

// GoogleCloudStorageConnection manages a connection to the Google Cloud
// Storage.
type GoogleCloudStorageConnection struct {
	GoogleOptions

	client *storage.Client
}

// NewDefaultGoogleCloudStorageConnection returns GoogleCloudStorageConnection
func NewDefaultGoogleCloudStorageConnection() *GoogleCloudStorageConnection {
	return &GoogleCloudStorageConnection{
		GoogleOptions: GoogleOptions{},
	}
}

// Open the connection to the Google Cloud Storage service.
// Creates a client with the given options.
// The context only needs to last until this function returns.
func (conn *GoogleCloudStorageConnection) Open(ctx context.Context) (*storage.Client, error) {
	if conn.client != nil {
		return conn.client, nil
	}
	client, err := storage.NewClient(ctx, conn.GetClientOptions()...)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to create storage client")
	}
	conn.client = client
	return client, nil
}

// Open a connection to the bucket that closes when ctx is done, and return an
// ObjectHandle pointing to path.
func (conn *GoogleCloudStorageConnection) getRemoteObject(ctx context.Context, path string) (*storage.ObjectHandle, error) {
	uri, err := parseURI(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	bucketName := uri.Hostname()
	bucket, err := conn.getRemoteBucket(ctx, bucketName)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	objectName := uri.Path[1:]
	logger.KV(xlog.DEBUG, "bucket", bucketName, "object", objectName)
	return bucket.Object(objectName), nil
}

func parseURI(path string) (*url.URL, error) {
	uri, err := url.Parse(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if uri.Scheme != "gs" {
		return nil, errors.Errorf("invalid scheme: %q", uri.Scheme)
	}
	if !strings.HasPrefix(uri.Path, "/") {
		return nil, errors.Errorf("invalid path: must start with /: got %s", uri.Path)
	}
	return uri, nil
}

func (conn *GoogleCloudStorageConnection) getRemoteBucket(ctx context.Context, bucketName string) (*storage.BucketHandle, error) {
	client, err := conn.Open(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to open storage client")
	}
	return client.Bucket(bucketName), nil
}

// GetReader returns a reader tied to ctx for path. The caller is responsible
// for calling Close on the reader when done.
func (conn *GoogleCloudStorageConnection) GetReader(ctx context.Context, path string) (result io.ReadCloser, err error) {
	obj, err := conn.getRemoteObject(ctx, path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to open reader")
	}
	return reader, nil
}

// GetWriter returns a writer tied to ctx for path.
// The caller is responsible for calling Close on the writer when done.
func (conn *GoogleCloudStorageConnection) GetWriter(ctx context.Context, path string) (result io.WriteCloser, err error) {
	obj, err := conn.getRemoteObject(ctx, path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return obj.NewWriter(ctx), nil
}

// SetContentType updates object with content type, if supported
func (conn *GoogleCloudStorageConnection) SetContentType(ctx context.Context, path, contentType string) error {
	obj, err := conn.getRemoteObject(ctx, path)
	if err != nil {
		return errors.WithStack(err)
	}
	attrs, err := obj.Update(ctx, storage.ObjectAttrsToUpdate{
		ContentType: contentType,
	})
	if err != nil {
		return errors.WithStack(err)
	}
	logger.ContextKV(ctx, xlog.DEBUG, "attrs", attrs)

	return nil
}

// Delete the file pointed to by path.
func (conn *GoogleCloudStorageConnection) Delete(ctx context.Context, path string) (err error) {
	obj, err := conn.getRemoteObject(ctx, path)
	if err != nil {
		return errors.WithStack(err)
	}
	if err := obj.Delete(ctx); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// List objects in the bucket, based on the path prefix.
// The returned iterator is not safe for concurrent operations without explicit synchronization.
// This wraps the bucket.Objects call and getting the bucket, to list the file objects
// If deliminator is not "", will make a "Delimiter" query
func (conn *GoogleCloudStorageConnection) List(ctx context.Context, path string, delimiter string) (*storage.ObjectIterator, error) {
	uri, err := parseURI(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	bucketName := uri.Hostname()
	bucket, err := conn.getRemoteBucket(ctx, bucketName)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	queryPath := uri.Path[1:]
	query := &storage.Query{
		// filter returned names on matching the prefix
		Prefix: queryPath,
		// indicates whether multiple versions of the same
		// object will be included in the results.
		Versions: false,
	}
	if delimiter != "" {
		query.Delimiter = delimiter
	}
	// Name, Metadata, Created
	//attrs := []string{"Name"} TODO This as a performance optimization
	//query.SetAttrSelection(attrs)
	// err = query.SetAttrSelection(attrs)
	logger.KV(xlog.DEBUG, "path", path)
	return bucket.Objects(ctx, query), nil
}

// Close the cloud storage connection, and any remaining open connections.
func (conn *GoogleCloudStorageConnection) Close() error {
	if conn.client != nil {
		if err := conn.client.Close(); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// Wait can be used to block on the completion of a write operation.
func (conn *GoogleCloudStorageConnection) Wait() error {
	return nil
}
