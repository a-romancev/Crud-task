package company

import "context"

type Repo interface {
	Create(ctx context.Context, request Company) (Company, error)
	Fetch(ctx context.Context, lookup Lookup) ([]Company, error)
	FetchOne(ctx context.Context, lookup Lookup) (Company, error)
	UpdateOne(ctx context.Context, lookup Lookup, request Company) (Company, error)
	DeleteOne(ctx context.Context, lookup Lookup) error
}

type CRUD struct {
	Repo Repo
}

func NewCRUD(repo Repo) *CRUD {
	return &CRUD{Repo: repo}
}
