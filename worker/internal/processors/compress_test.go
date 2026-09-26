package processors

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"parcel/worker/internal/models"
)

type fakeS3Put struct {
	called bool
	key    string
	body   []byte
}

func (f *fakeS3Put) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	f.called = true
	f.key = aws.ToString(params.Key)
	f.body, _ = io.ReadAll(params.Body)
	return &s3.PutObjectOutput{}, nil
}

func TestCompressProcessorCompressesContent(t *testing.T) {
	raw := bytes.Repeat([]byte("parcel compress test "), 500)

	result, err := CompressProcessor{}.Run(context.Background(), fakeS3{body: raw}, models.Job{
		Bucket: "parcel-files", Key: "uploads/f/test.txt", Operation: "compress",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Size >= int64(len(raw)) {
		t.Errorf("compressed size %d should be less than original %d", result.Size, len(raw))
	}
	if result.ContentType != "application/gzip" {
		t.Errorf("ContentType = %s, want application/gzip", result.ContentType)
	}
	if result.SHA256 == "" {
		t.Error("SHA256 is empty")
	}
}

func TestCompressProcessorRejectsEmptyFile(t *testing.T) {
	_, err := CompressProcessor{}.Run(context.Background(), fakeS3{body: []byte{}}, models.Job{
		Bucket: "parcel-files", Key: "uploads/f/empty.bin", Operation: "compress",
	})
	if err == nil {
		t.Fatal("Run() error = nil, want error for empty file")
	}
}

func TestCompressProcessorUploadsToProcessed(t *testing.T) {
	raw := bytes.Repeat([]byte("x"), 100)
	put := &fakeS3Put{}

	cp := CompressProcessor{PutClient: put}
	_, err := cp.Run(context.Background(), fakeS3{body: raw}, models.Job{
		Bucket: "parcel-files", Key: "uploads/f/test.bin", Operation: "compress",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !put.called {
		t.Fatal("PutObject was not called")
	}
	if put.key != "processed/f/test.gz" {
		t.Errorf("PutObject key = %s, want processed/f/test.gz", put.key)
	}
}

func TestCompressProcessorPropagatesS3Error(t *testing.T) {
	_, err := CompressProcessor{}.Run(context.Background(), fakeS3{err: io.ErrUnexpectedEOF}, models.Job{
		Bucket: "parcel-files", Key: "uploads/f/missing.bin", Operation: "compress",
	})
	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}
}

func TestCompressOutputIsDecompressable(t *testing.T) {
	raw := []byte("decompress me after gzip")
	put := &fakeS3Put{}

	cp := CompressProcessor{PutClient: put}
	_, err := cp.Run(context.Background(), fakeS3{body: raw}, models.Job{
		Bucket: "parcel-files", Key: "uploads/f/msg.txt", Operation: "compress",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	gz, err := gzip.NewReader(bytes.NewReader(put.body))
	if err != nil {
		t.Fatalf("gzip.NewReader error = %v", err)
	}
	defer gz.Close()

	decompressed, err := io.ReadAll(gz)
	if err != nil {
		t.Fatalf("ReadAll error = %v", err)
	}
	if !bytes.Equal(decompressed, raw) {
		t.Errorf("decompressed content mismatch: got %q, want %q", decompressed, raw)
	}
}

func TestDispatchReturnsCompressProcessor(t *testing.T) {
	processor, err := Dispatch("compress")
	if err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}
	if _, ok := processor.(CompressProcessor); !ok {
		t.Fatalf("Dispatch() = %T, want CompressProcessor", processor)
	}
}