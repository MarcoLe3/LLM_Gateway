package repo

import (
	"context"
)

type Repository interface {
	Insert(ctx context.Context) error
	
}