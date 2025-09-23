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

type Processor struct {
	client   *llm.GeminiClient
	database db.Database
	logger   *slog.Logger
}

func (p *Processor) Run(ctx context.Context, ids <-chan int64) {
	for id := range ids {
		err := p.handleIDWithRetry(ctx, id)
		if err != nil {
			p.logger.ErrorContext(ctx, "Failed to handle ID", "id", id, "error", err)
			dbErr := p.database.Queries().UpdateHNComment(ctx, queries.UpdateHNCommentParams{ID: id, Status: "failed"})
			if dbErr != nil {
				p.logger.ErrorContext(ctx, "Failed to update HN comment to errored", "id", id, "error", err)
			}
		}
	}
}

const maxAttempts = 5

func (p *Processor) handleIDWithRetry(ctx context.Context, id int64) error {
	for attempt := range maxAttempts {
		err := p.handleID(ctx, id)
		if err == nil {
			return nil
		} else if errors.Is(err, llm.ErrQuotaExhausted) {
			p.logger.DebugContext(ctx, "Quota exhaused", "id", id)
			select {
			case <-time.After(24 * time.Hour):
			case <-ctx.Done():
				return ctx.Err()
			}
		} else if errors.Is(err, llm.ErrRateLimit) &&
			errors.Is(err, llm.ErrNoResponse) &&
			errors.Is(err, llm.ErrServiceUnavailable) &&
			errors.Is(err, llm.ErrQuotaExhausted) {
			delay := calculateBackoffDelay(attempt)
			p.logger.DebugContext(ctx, "Retrying ID", "id", id, "delay", delay)

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		} else {
			return err
		}
	}
	return errors.New("max attempts exceeded")
}

func (p *Processor) handleID(ctx context.Context, id int64) error {
	p.logger.DebugContext(ctx, "Handling ID", "id", id)
	err := p.database.Queries().UpdateHNComment(ctx, queries.UpdateHNCommentParams{
		Status: "in_progress",
		ID:     id,
	})
	if err != nil {
		return err
	}

	value, err := p.database.Queries().GetHNCommentValue(ctx, id)
	if err != nil {
		return err
	}

	p.logger.DebugContext(ctx, "Parsing value", "value", value)
	jobPosting, err := p.client.ParseJobPosting(ctx, value)
	if err != nil {
		return err
	}

	p.logger.DebugContext(ctx, "Handling parsed job data", "data", jobPosting)
	err = p.insertJobs(ctx, id, jobPosting)
	if err != nil {
		return err
	}

	p.logger.DebugContext(ctx, "Completed handling ID", "id", id)
	return p.database.Queries().UpdateHNComment(ctx, queries.UpdateHNCommentParams{
		Status: "completed",
		ID:     id,
	})
}

func (p *Processor) insertJobs(ctx context.Context, id int64, jobPosting llm.JobPosting) error {
	tx, err := p.database.DB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := queries.New(tx)

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
			HnCommentID:        id,
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
	jitter := (rand.Float64() - 0.5) * 2 * jitterRange

	return max(time.Duration(float64(delay)+jitter), baseDelay)
}
