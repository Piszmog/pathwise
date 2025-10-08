package llm

import "context"

type Client interface {
	ParseJobPostings(ctx context.Context, inputs map[int64]string) ([]JobPosting, error)
	Close() error
}
