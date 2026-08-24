package models

import "testing"

func TestDecodeJobParsesValidBody(t *testing.T) {
	body := `{"job_id":"job1","file_id":"file1","bucket":"parcel-files","key":"uploads/file1/photo.jpg","operation":"inspect"}`

	job, err := DecodeJob(body)
	if err != nil {
		t.Fatalf("DecodeJob() error = %v", err)
	}

	want := Job{JobID: "job1", FileID: "file1", Bucket: "parcel-files", Key: "uploads/file1/photo.jpg", Operation: "inspect"}
	if job != want {
		t.Fatalf("DecodeJob() = %+v, want %+v", job, want)
	}
}

func TestDecodeJobRejectsInvalidJSON(t *testing.T) {
	if _, err := DecodeJob("not json"); err == nil {
		t.Fatal("DecodeJob() error = nil, want error")
	}
}

func TestDecodeJobRejectsMissingFields(t *testing.T) {
	cases := []string{
		`{"file_id":"file1","bucket":"b","key":"k","operation":"inspect"}`,
		`{"job_id":"job1","bucket":"b","key":"k","operation":"inspect"}`,
		`{"job_id":"job1","file_id":"file1","key":"k","operation":"inspect"}`,
		`{"job_id":"job1","file_id":"file1","bucket":"b","operation":"inspect"}`,
		`{"job_id":"job1","file_id":"file1","bucket":"b","key":"k"}`,
	}

	for _, body := range cases {
		if _, err := DecodeJob(body); err == nil {
			t.Fatalf("DecodeJob(%q) error = nil, want error", body)
		}
	}
}
