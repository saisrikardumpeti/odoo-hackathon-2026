package asset_repo

import "github.com/jackc/pgx/v5/pgxpool"

type AssetRepository struct {
	pool *pgxpool.Pool
}

func NewAssetRepository(pool *pgxpool.Pool) *AssetRepository {
	return &AssetRepository{pool: pool}
}