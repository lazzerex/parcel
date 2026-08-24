package models

import (
	"encoding/json"
	"fmt"
)

type Job struct {
	JobID     string `json:"job_id"`
	FileID    string `json:"file_id"`
	Bucket    string `json:"bucket"`
	Key       string `json:"key"`
	Operation string `json:"operation"`
}

func DecodeJob(body string) (Job, error) {
	var job Job
	if err := json.Unmarshal([]byte(body), &job); err != nil {
		return Job{}, fmt.Errorf("decode job: %w", err)
	}

	if job.JobID == "" || job.FileID == "" || job.Bucket == "" || job.Key == "" || job.Operation == "" {
		return Job{}, fmt.Errorf("job missing required field: %+v", job)
	}

	return job, nil
}
