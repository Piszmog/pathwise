package llm

import "context"

type Client interface {
	ParseJobPosting(ctx context.Context, input string) (JobPosting, error)
	Close() error
}
