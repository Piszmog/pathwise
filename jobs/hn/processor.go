package hn

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"math"
	"math/rand/v2"
	"time"

	"github.com/Piszmog/pathwise/internal/db"
	"github.com/Piszmog/pathwise/internal/db/queries"
	"github.com/Piszmog/pathwise/jobs/llm"
	"github.com/google/uuid"
)

const (
	maxAttempts  = 5
	batchSize    = 10
	batchTimeout = 30 * time.Second
)

type Processor struct {
	client   *llm.GeminiClient
	database db.Database
	logger   *slog.Logger
}

func NewProcessor(logger *slog.Logger, database db.Database, client *llm.GeminiClient) *Processor {
	return &Processor{
		client:   client,
		database: database,
		logger:   logger,
	}
}

func (p *Processor) Run(ctx context.Context, ids <-chan int64) {
	for {
		batch := p.collectBatch(ctx, ids)
		if len(batch) == 0 {
			return
		}

		err := p.handleBatchWithRetry(ctx, batch)
		if err != nil {
			p.logger.ErrorContext(ctx, "failed to handle batch", "ids", batch, "error", err)
			dbErr := p.database.Queries().UpdateHNComments(ctx, queries.UpdateHNCommentsParams{Ids: batch, Status: "failed"})
			if dbErr != nil {
				p.logger.ErrorContext(ctx, "failed to update HN comments to failed", "ids", batch, "error", dbErr)
			}
		}
	}
}

func (p *Processor) collectBatch(ctx context.Context, ids <-chan int64) []int64 {
	var batch []int64
	for {
		timeout := time.After(batchTimeout)

		select {
		case id, ok := <-ids:
			if !ok {
				return batch
			}
			batch = append(batch, id)
			if len(batch) >= batchSize {
				return batch
			}
		case <-timeout:
			if len(batch) > 0 {
				return batch
			}
		case <-ctx.Done():
			return batch
		}
	}
}

func (p *Processor) handleBatchWithRetry(ctx context.Context, ids []int64) error {
	for attempt := range maxAttempts {
		err := p.handleIDs(ctx, ids)
		switch {
		case err == nil:
			p.logger.DebugContext(ctx, "completed handling IDs", "ids", ids)
			return p.database.Queries().UpdateHNComments(ctx, queries.UpdateHNCommentsParams{
				Status: "completed",
				Ids:    ids,
			})
		case errors.Is(err, llm.ErrQuotaExhausted):
			p.logger.DebugContext(ctx, "quota exhausted", "ids", ids, "error", err)
			select {
			case <-time.After(24 * time.Hour):
			case <-ctx.Done():
				return ctx.Err()
			}
		case errors.Is(err, llm.ErrRateLimit) ||
			errors.Is(err, llm.ErrNoResponse) ||
			errors.Is(err, llm.ErrServiceUnavailable):
			delay := calculateBackoffDelay(attempt)
			p.logger.DebugContext(ctx, "retrying IDs", "ids", ids, "delay", delay, "error", err)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		default:
			return err
		}
	}
	return errMaxAttempts
}

var errMaxAttempts = errors.New("max attempts exceeded")

func (p *Processor) handleIDs(ctx context.Context, ids []int64) error {
	p.logger.DebugContext(ctx, "handling IDs", "ids", ids)
	err := p.database.Queries().UpdateHNComments(ctx, queries.UpdateHNCommentsParams{
		Status: "in_progress",
		Ids:    ids,
	})
	if err != nil {
		return err
	}

	values, err := p.database.Queries().GetHNCommentValues(ctx, ids)
	if err != nil {
		return err
	}

	if len(values) == 0 {
		return nil
	}

	valuesToParse := make(map[int64]string)
	for _, row := range values {
		if row.Value == "" {
			continue
		}
		valuesToParse[row.ID] = row.Value
	}

	p.logger.DebugContext(ctx, "parsing values", "values", valuesToParse)
	jobPostings, err := p.client.ParseJobPostings(ctx, valuesToParse)
	if err != nil {
		return err
	}

	p.logger.DebugContext(ctx, "handling parsed job data", "data", jobPostings)
	return p.insertJobs(ctx, jobPostings)
}

func (p *Processor) insertJobs(ctx context.Context, jobPostings []llm.JobPosting) error {
	tx, err := p.database.DB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if txErr := tx.Rollback(); txErr != nil {
			err = errors.Join(err, txErr)
		}
	}()

	q := queries.New(tx)

	for _, jobPosting := range jobPostings {
		p.logger.DebugContext(ctx, "inserting job posting data", "id", jobPosting.ID, "data", jobPosting)
		for _, job := range jobPosting.Jobs {
			jobID := uuid.NewString()
			salary := job.Compensation.BaseSalary
			if salary == "" {
				salary = jobPosting.GeneralCompensation.BaseSalary
			}
			equity := job.Compensation.Equity
			if equity == "" {
				equity = jobPosting.GeneralCompensation.Equity
			}
			jobParam := queries.InsertHNJobParams{
				Company:            jobPosting.CompanyName,
				CompanyDescription: jobPosting.CompanyDescription,
				Title:              job.Title,
				ID:                 jobID,
				CompanyUrl:         newNullString(jobPosting.CompanyURL),
				ContactEmail:       newNullString(jobPosting.ContactEmail),
				Description:        newNullString(job.Description),
				RoleType:           newNullString(job.RoleType),
				Location:           newNullString(jobPosting.Location),
				Salary:             newNullString(salary),
				Equity:             newNullString(equity),
				IsHybrid:           boolToInt64(jobPosting.IsHybrid),
				IsRemote:           boolToInt64(jobPosting.IsRemote),
				HnCommentID:        jobPosting.ID,
			}

			err = q.InsertHNJob(ctx, jobParam)
			if err != nil {
				return err
			}

			for _, tech := range job.TechStack {
				stack := queries.InsertHNTechStackParams{
					HnJobID: jobID,
					Value:   tech,
				}

				err = q.InsertHNTechStack(ctx, stack)
				if err != nil {
					return err
				}
			}
		}
	}

	return tx.Commit()
}

func boolToInt64(b bool) int64 {
	var i int64
	if b {
		i = 1
	} else {
		i = 0
	}
	return i
}

func newNullString(value string) sql.NullString {
	var isValid bool
	if value != "" {
		isValid = true
	}
	return sql.NullString{
		Valid:  isValid,
		String: value,
	}
}

func calculateBackoffDelay(attempt int) time.Duration {
	baseDelay := 4 * time.Second
	maxDelay := 5 * time.Minute

	delay := min(time.Duration(float64(baseDelay)*math.Pow(2, float64(attempt))), maxDelay)

	jitterRange := float64(delay) * 0.25
	// #nosec G404 -- Using weak RNG for jitter is acceptable
	jitter := (rand.Float64() - 0.5) * 2 * jitterRange

	return max(time.Duration(float64(delay)+jitter), baseDelay)
}
