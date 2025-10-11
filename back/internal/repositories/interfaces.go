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
	// パラメータで検索（完全一致）
	SearchByParameters(ctx context.Context, userID int64, subject string, prompt string, filters map[string]interface{}) ([]*models.Problem, error)
	// フリーワード検索（部分一致）
	SearchByKeyword(ctx context.Context, userID int64, keyword string, limit, offset int) ([]*models.Problem, error)
}

type SessionRepository interface {
	Create(ctx context.Context, session *models.Session) error
	GetByToken(ctx context.Context, token string) (*models.Session, error)
	Delete(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
}
