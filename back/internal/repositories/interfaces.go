package repositories

import (
	"context"
	"github.com/mon-gene/back/internal/models"
)

type UserRepository interface {
	GetBySchoolCode(ctx context.Context, schoolCode string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
}

type ProblemRepository interface {
	Create(ctx context.Context, problem *models.Problem) error
	GetByID(ctx context.Context, id int64) (*models.Problem, error)
	GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*models.Problem, error)
	Delete(ctx context.Context, id int64) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *models.Session) error
	GetByToken(ctx context.Context, token string) (*models.Session, error)
	Delete(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
}
