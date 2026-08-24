package processors

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"

	"parcel/worker/internal/models"
)

type fakeS3 struct {
	body []byte
	err  error
}

func (f fakeS3) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(f.body))}, nil
}

func TestInspectProcessorComputesSizeAndHash(t *testing.T) {
	content := bytes.Repeat([]byte("parcel"), 200)
	sum := sha256.Sum256(content)

	result, err := InspectProcessor{}.Run(context.Background(), fakeS3{body: content}, models.Job{
		Bucket: "parcel-files", Key: "uploads/f/x.bin", Operation: "inspect",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", result.Size, len(content))
	}
	if result.SHA256 != hex.EncodeToString(sum[:]) {
		t.Errorf("SHA256 = %s, want %s", result.SHA256, hex.EncodeToString(sum[:]))
	}
}

func TestInspectProcessorDetectsContentType(t *testing.T) {
	png := []byte("\x89PNG\r\n\x1a\n" + string(bytes.Repeat([]byte{0}, 100)))

	result, err := InspectProcessor{}.Run(context.Background(), fakeS3{body: png}, models.Job{
		Bucket: "parcel-files", Key: "uploads/f/x.png", Operation: "inspect",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.ContentType != "image/png" {
		t.Errorf("ContentType = %s, want image/png", result.ContentType)
	}
}

func TestInspectProcessorHandlesEmptyObject(t *testing.T) {
	sum := sha256.Sum256(nil)

	result, err := InspectProcessor{}.Run(context.Background(), fakeS3{body: []byte{}}, models.Job{
		Bucket: "parcel-files", Key: "uploads/f/empty.bin", Operation: "inspect",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.Size != 0 {
		t.Errorf("Size = %d, want 0", result.Size)
	}
	if result.SHA256 != hex.EncodeToString(sum[:]) {
		t.Errorf("SHA256 = %s, want %s", result.SHA256, hex.EncodeToString(sum[:]))
	}
}

func TestInspectProcessorPropagatesS3Error(t *testing.T) {
	_, err := InspectProcessor{}.Run(context.Background(), fakeS3{err: errors.New("no such key")}, models.Job{
		Bucket: "parcel-files", Key: "uploads/f/missing.bin", Operation: "inspect",
	})
	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}
}

func TestDispatchReturnsInspectProcessor(t *testing.T) {
	processor, err := Dispatch("inspect")
	if err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}
	if _, ok := processor.(InspectProcessor); !ok {
		t.Fatalf("Dispatch() = %T, want InspectProcessor", processor)
	}
}

func TestDispatchRejectsUnknownOperation(t *testing.T) {
	if _, err := Dispatch("compress"); err == nil {
		t.Fatal("Dispatch() error = nil, want error")
	}
}
