package example_repo

import "database/sql"

type ExampleRepository struct {
	db *sql.DB
}

func NewExampleRepository(db *sql.DB) *ExampleRepository {
	return &ExampleRepository{
		db: db,
	}
}
