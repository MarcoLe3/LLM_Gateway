package worker_test

import (
	"llm_gateway/internal/job"
	"testing"

	"github.com/google/uuid"
)

func TestProcessJob(t *testing.T) {
	id := uuid.New()
	j := &job.Job{ID: id, Attempts: 1}

	t.Run("success marks succeeded", func(t *testing.T){
		
	})
}
