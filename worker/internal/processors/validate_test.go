package processors

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"parcel/worker/internal/models"
)

func TestValidateProcessorPassesOnValidFile(t *testing.T) {
	content := bytes.Repeat([]byte("hello parcel"), 100)
	sum := sha256.Sum256(content)

	result, err := ValidateProcessor{}.Run(context.Background(), fakeS3{body: content}, models.Job{
		Bucket: "parcel-files", Key: "uploads/f/valid.txt", Operation: "validate",
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

func TestValidateProcessorRejectsEmptyFile(t *testing.T) {
	_, err := ValidateProcessor{}.Run(context.Background(), fakeS3{body: []byte{}}, models.Job{
		Bucket: "parcel-files", Key: "uploads/f/empty.bin", Operation: "validate",
	})
	if err == nil {
		t.Fatal("Run() error = nil, want error for empty file")
	}
}

func TestValidateProcessorRejectsOversizedFile(t *testing.T) {
	content := bytes.Repeat([]byte("x"), maxFileSize+1)

	_, err := ValidateProcessor{}.Run(context.Background(), fakeS3{body: content}, models.Job{
		Bucket: "parcel-files", Key: "uploads/f/huge.bin", Operation: "validate",
	})
	if err == nil {
		t.Fatal("Run() error = nil, want error for oversized file")
	}
}

func TestValidateProcessorRejectsDisallowedContentType(t *testing.T) {
	content := append([]byte("\x7fELF"), bytes.Repeat([]byte{0}, 100)...)

	_, err := ValidateProcessor{}.Run(context.Background(), fakeS3{body: content}, models.Job{
		Bucket: "parcel-files", Key: "uploads/f/binary", Operation: "validate",
	})
	if err == nil {
		t.Fatal("Run() error = nil, want error for disallowed content type")
	}
}

func TestValidateProcessorPropagatesS3Error(t *testing.T) {
	_, err := ValidateProcessor{}.Run(context.Background(), fakeS3{err: fmt.Errorf("access denied")}, models.Job{
		Bucket: "parcel-files", Key: "uploads/f/missing.bin", Operation: "validate",
	})
	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}
}

func TestDispatchReturnsValidateProcessor(t *testing.T) {
	processor, err := Dispatch("validate")
	if err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}
	if _, ok := processor.(ValidateProcessor); !ok {
		t.Fatalf("Dispatch() = %T, want ValidateProcessor", processor)
	}
}
