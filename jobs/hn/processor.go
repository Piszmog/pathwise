package hn

import (
	"context"
	"database/sql"
	"log/slog"

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

func (p *Processor) Run(ctx context.Context, ids <-chan int64) error {
	for id := range ids {
		p.handleID(ctx, id)
	}

	return nil
}

func (p *Processor) handleID(ctx context.Context, id int64) error {
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

	jobPosting, err := p.client.ParseJobPosting(ctx, value)
	if err != nil {
		return err
	}

	err = p.insertJobs(ctx, id, jobPosting)
	if err != nil {
		return err
	}

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
