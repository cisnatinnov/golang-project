// This file contains the interfaces for the repository layer.
// The repository layer is responsible for interacting with the database.
// For testing purpose we will generate mock implementations of these
// interfaces using mockgen. See the Makefile for more information.
package repository

import "context"

type RepositoryInterface interface {
	CreateEstate(ctx context.Context, input CreateEstateInput) (output CreateEstateOutput, err error)
	GetEstateById(ctx context.Context, id string) (estate Estate, err error)
	CreateTree(ctx context.Context, input CreateTreeInput) (output CreateTreeOutput, err error)
	GetEstateStats(ctx context.Context, input GetEstateStatsInput) (output GetEstateStatsOutput, err error)
	GetTreesByEstateId(ctx context.Context, input GetTreesByEstateIdInput) (output GetTreesByEstateIdOutput, err error)
}
