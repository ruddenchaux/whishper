package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/pluja/anysub/db"
	"github.com/pluja/anysub/ent"
	"github.com/pluja/anysub/models"
)

const (
	TypeTranscription = "transcription:create"

	// CPU transcription with diarization runs slower than real time, so
	// asynq's 30m default timeout fires mid-job. The handler doesn't watch
	// ctx, so the timed-out run keeps going while asynq schedules a retry,
	// stacking duplicate whisperx runs on the same file.
	transcriptionTimeout  = 12 * time.Hour
	transcriptionMaxRetry = 1
)

func NewTranscriptionTask(tx ent.Transcription) (*asynq.Task, error) {
	payload, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeTranscription, payload,
		asynq.Timeout(transcriptionTimeout),
		asynq.MaxRetry(transcriptionMaxRetry),
	), nil
}

func HandleNewTranscriptionTask(ctx context.Context, t *asynq.Task) error {
	var tx ent.Transcription
	if err := json.Unmarshal(t.Payload(), &tx); err != nil {
		db.Client().Transcription.UpdateOneID(tx.ID).SetStatus(models.TsStatusError).Save(context.Background())
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	err := transcribe(&tx)
	if err != nil {
		db.Client().Transcription.UpdateOneID(tx.ID).SetStatus(models.TsStatusError).Save(context.Background())
		return fmt.Errorf("transcription failed failed: %v", err)
	}

	return nil
}
