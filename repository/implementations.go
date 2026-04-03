package repository

import (
	"context"

	"github.com/google/uuid"
)

func (r *Repository) CreateEstate(ctx context.Context, input CreateEstateInput) (output CreateEstateOutput, err error) {
	id := uuid.New().String()
	_, err = r.Db.ExecContext(ctx, "INSERT INTO estate (id, length, width) VALUES ($1, $2, $3)", id, input.Length, input.Width)
	if err != nil {
		return
	}
	output.Id = id
	return
}

func (r *Repository) GetEstateById(ctx context.Context, id string) (estate Estate, err error) {
	err = r.Db.QueryRowContext(ctx, "SELECT id, length, width FROM estate WHERE id = $1", id).Scan(&estate.Id, &estate.Length, &estate.Width)
	return
}

func (r *Repository) CreateTree(ctx context.Context, input CreateTreeInput) (output CreateTreeOutput, err error) {
	id := uuid.New().String()
	_, err = r.Db.ExecContext(ctx, "INSERT INTO tree (id, estate_id, x, y, height) VALUES ($1, $2, $3, $4, $5)", id, input.EstateId, input.X, input.Y, input.Height)
	if err != nil {
		return
	}
	output.Id = id
	return
}

func (r *Repository) GetEstateStats(ctx context.Context, input GetEstateStatsInput) (output GetEstateStatsOutput, err error) {
	err = r.Db.QueryRowContext(ctx, `
		SELECT 
			COUNT(id), 
			COALESCE(MAX(height), 0), 
			COALESCE(MIN(height), 0), 
			COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY height), 0)
		FROM tree 
		WHERE estate_id = $1
	`, input.EstateId).Scan(&output.Count, &output.Max, &output.Min, &output.Median)
	return
}

func (r *Repository) GetTreesByEstateId(ctx context.Context, input GetTreesByEstateIdInput) (output GetTreesByEstateIdOutput, err error) {
	rows, err := r.Db.QueryContext(ctx, "SELECT id, estate_id, x, y, height FROM tree WHERE estate_id = $1", input.EstateId)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var t Tree
		if err = rows.Scan(&t.Id, &t.EstateId, &t.X, &t.Y, &t.Height); err != nil {
			return
		}
		output.Trees = append(output.Trees, t)
	}
	err = rows.Err()
	return
}
