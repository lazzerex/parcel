package processors

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

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

type sequencedS3 struct {
	responses [][]byte
	errs      []error
	calls     int
}

func (f *sequencedS3) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	i := f.calls
	f.calls++
	if f.errs[i] != nil {
		return nil, f.errs[i]
	}
	return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(f.responses[i]))}, nil
}

func TestInspectProcessorRetriesOnNoSuchKeyThenSucceeds(t *testing.T) {
	content := []byte("parcel test content")
	sum := sha256.Sum256(content)
	fake := &sequencedS3{
		errs:      []error{&types.NoSuchKey{}, nil},
		responses: [][]byte{nil, content},
	}

	result, err := InspectProcessor{}.Run(context.Background(), fake, models.Job{
		Bucket: "parcel-files", Key: "uploads/f/x.bin", Operation: "inspect",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if fake.calls != 2 {
		t.Errorf("GetObject called %d times, want 2", fake.calls)
	}
	if result.SHA256 != hex.EncodeToString(sum[:]) {
		t.Errorf("SHA256 = %s, want %s", result.SHA256, hex.EncodeToString(sum[:]))
	}
}

func TestInspectProcessorGivesUpAfterExhaustingRetries(t *testing.T) {
	fake := &sequencedS3{
		errs:      []error{&types.NoSuchKey{}, &types.NoSuchKey{}, &types.NoSuchKey{}, &types.NoSuchKey{}},
		responses: [][]byte{nil, nil, nil, nil},
	}

	_, err := InspectProcessor{}.Run(context.Background(), fake, models.Job{
		Bucket: "parcel-files", Key: "uploads/f/x.bin", Operation: "inspect",
	})
	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}
	if fake.calls != len(objectNotFoundRetryDelays)+1 {
		t.Errorf("GetObject called %d times, want %d", fake.calls, len(objectNotFoundRetryDelays)+1)
	}
}

func TestInspectProcessorDoesNotRetryOtherErrors(t *testing.T) {
	fake := &sequencedS3{
		errs:      []error{errors.New("access denied")},
		responses: [][]byte{nil},
	}

	_, err := InspectProcessor{}.Run(context.Background(), fake, models.Job{
		Bucket: "parcel-files", Key: "uploads/f/x.bin", Operation: "inspect",
	})
	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}
	if fake.calls != 1 {
		t.Errorf("GetObject called %d times, want 1 (non-NoSuchKey errors must not retry)", fake.calls)
	}
}

func TestGetObjectWithRetryRespectsContextCancellation(t *testing.T) {
	fake := &sequencedS3{
		errs:      []error{&types.NoSuchKey{}, &types.NoSuchKey{}},
		responses: [][]byte{nil, nil},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := getObjectWithRetry(ctx, fake, models.Job{Bucket: "parcel-files", Key: "uploads/f/x.bin"})
	if err == nil {
		t.Fatal("getObjectWithRetry() error = nil, want context deadline error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error = %v, want context.DeadlineExceeded", err)
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
