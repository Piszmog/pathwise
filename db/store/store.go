package store

import "context"

type Store[T Record] interface {
	GetByID(ctx context.Context, id int) (T, error)
	Get(ctx context.Context, opts GetOpts) ([]T, error)
	GetAllByID(ctx context.Context, id int) ([]T, error)
	Insert(ctx context.Context, rec T) error
	Update(ctx context.Context, rec T) error
	Delete(ctx context.Context, id int) error
}

type GetOpts struct {
	Page    int
	PerPage int
}

type Record interface {
	RecordID() int
}
