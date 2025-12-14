package services

import (
	"context"
	"fmt"
	"time"

	"github.com/mon-gene/back/internal/models"
	"github.com/mon-gene/back/internal/repositories"
)

type SearchFilterService interface {
	CreateSearchFilter(ctx context.Context, req models.CreateSearchFilterRequest, userID int64) (*models.SearchFilter, error)
	GetSearchFilter(ctx context.Context, id int64, userID int64) (*models.SearchFilter, error)
	GetUserSearchFilters(ctx context.Context, userID int64) ([]*models.SearchFilter, error)
	UpdateSearchFilter(ctx context.Context, req models.UpdateSearchFilterRequest, userID int64) (*models.SearchFilter, error)
	DeleteSearchFilter(ctx context.Context, id int64, userID int64) error
}

type searchFilterService struct {
	searchFilterRepo repositories.SearchFilterRepository
}

func NewSearchFilterService(searchFilterRepo repositories.SearchFilterRepository) SearchFilterService {
	return &searchFilterService{
		searchFilterRepo: searchFilterRepo,
	}
}

func (s *searchFilterService) CreateSearchFilter(ctx context.Context, req models.CreateSearchFilterRequest, userID int64) (*models.SearchFilter, error) {
	// バリデーション
	if req.Name == "" {
		return nil, fmt.Errorf("検索条件の名前は必須です")
	}

	filter := &models.SearchFilter{
		UserID:      userID,
		Name:        req.Name,
		Keyword:     req.Keyword,
		Subject:     req.Subject,
		Units:       req.Units,
		Year:        req.Year,
		ExamSession: req.ExamSession,
		IsChecked:   req.IsChecked,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.searchFilterRepo.Create(ctx, filter); err != nil {
		return nil, fmt.Errorf("検索条件の作成に失敗しました: %w", err)
	}

	return filter, nil
}

func (s *searchFilterService) GetSearchFilter(ctx context.Context, id int64, userID int64) (*models.SearchFilter, error) {
	filter, err := s.searchFilterRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// ユーザーの所有権を確認
	if filter.UserID != userID {
		return nil, fmt.Errorf("アクセス権限がありません")
	}

	return filter, nil
}

func (s *searchFilterService) GetUserSearchFilters(ctx context.Context, userID int64) ([]*models.SearchFilter, error) {
	filters, err := s.searchFilterRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("検索条件の取得に失敗しました: %w", err)
	}

	return filters, nil
}

func (s *searchFilterService) UpdateSearchFilter(ctx context.Context, req models.UpdateSearchFilterRequest, userID int64) (*models.SearchFilter, error) {
	// バリデーション
	if req.Name == "" {
		return nil, fmt.Errorf("検索条件の名前は必須です")
	}

	// 既存の検索条件を取得
	filter, err := s.searchFilterRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// ユーザーの所有権を確認
	if filter.UserID != userID {
		return nil, fmt.Errorf("アクセス権限がありません")
	}

	// 更新
	filter.Name = req.Name
	filter.Keyword = req.Keyword
	filter.Subject = req.Subject
	filter.Units = req.Units
	filter.Year = req.Year
	filter.ExamSession = req.ExamSession
	filter.IsChecked = req.IsChecked
	filter.UpdatedAt = time.Now()

	if err := s.searchFilterRepo.Update(ctx, filter); err != nil {
		return nil, fmt.Errorf("検索条件の更新に失敗しました: %w", err)
	}

	return filter, nil
}

func (s *searchFilterService) DeleteSearchFilter(ctx context.Context, id int64, userID int64) error {
	// 既存の検索条件を取得
	filter, err := s.searchFilterRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// ユーザーの所有権を確認
	if filter.UserID != userID {
		return fmt.Errorf("アクセス権限がありません")
	}

	if err := s.searchFilterRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("検索条件の削除に失敗しました: %w", err)
	}

	return nil
}