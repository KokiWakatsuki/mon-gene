package services

import (
	"errors"
	"github.com/mon-gene/back/internal/models"
	"github.com/mon-gene/back/internal/repositories"
)

type SourceListService struct {
	repo repositories.SourceListRepository
}

func NewSourceListService(repo repositories.SourceListRepository) *SourceListService {
	return &SourceListService{repo: repo}
}

// AddItem は出典リスト項目を追加
func (s *SourceListService) AddItem(userID int64, year string, examSession string) (*models.SourceListItem, error) {
	if year == "" || examSession == "" {
		return nil, errors.New("year and exam_session are required")
	}

	item := &models.SourceListItem{
		UserID:      userID,
		Year:        year,
		ExamSession: examSession,
	}

	err := s.repo.Create(item)
	if err != nil {
		return nil, err
	}

	return item, nil
}

// GetUserItems はユーザーの出典リストを取得
func (s *SourceListService) GetUserItems(userID int64) ([]*models.SourceListItem, error) {
	return s.repo.GetByUserID(userID)
}

// DeleteItem は出典リスト項目を削除
func (s *SourceListService) DeleteItem(id int64, userID int64) error {
	return s.repo.Delete(id, userID)
}

// DeleteByYearAndExam は年度と回数で出典リスト項目を削除
func (s *SourceListService) DeleteByYearAndExam(userID int64, year string, examSession string) error {
	return s.repo.DeleteByYearAndExam(userID, year, examSession)
}