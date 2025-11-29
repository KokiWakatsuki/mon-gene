package repositories

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/mon-gene/back/internal/models"
	"github.com/mon-gene/back/internal/utils"
)

type memoryUserRepository struct {
	users map[string]*models.User
	mutex sync.RWMutex
}

func NewMemoryUserRepository() UserRepository {
	repo := &memoryUserRepository{
		users: make(map[string]*models.User),
		mutex: sync.RWMutex{},
	}
	
	// seedデータを追加
	repo.seedData()
	
	return repo
}

func (r *memoryUserRepository) seedData() {
	// CSVファイルからユーザーデータを読み込み
	users, err := r.loadUsersFromCSV()
	if err != nil {
		log.Printf("⚠️ CSVファイルの読み込みに失敗しました: %v", err)
		log.Printf("📝 フォールバック: デフォルトユーザーを作成します")
		r.createDefaultUser()
		return
	}

	log.Printf("✅ CSVファイルから %d 人のユーザーを読み込みました", len(users))
	
	for _, user := range users {
		r.users[user.SchoolCode] = user
	}
}

func (r *memoryUserRepository) loadUsersFromCSV() ([]*models.User, error) {
	// CSVファイルのパスを取得
	csvPath := filepath.Join("data", "users.csv")
	
	// ファイルを開く
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("CSVファイルを開けません: %w", err)
	}
	defer file.Close()

	// CSVリーダーを作成
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("CSVファイルの読み込みに失敗しました: %w", err)
	}

	if len(records) < 2 { // ヘッダー + 最低1行のデータ
		return nil, fmt.Errorf("CSVファイルにデータがありません")
	}

	var users []*models.User
	now := time.Now()

	// ヘッダー行をスキップして処理
	for i, record := range records[1:] {
		if len(record) < 11 {
			log.Printf("⚠️ 行 %d: 列数が不足しています (期待値: 11, 実際: %d)", i+2, len(record))
			continue
		}

		// IDを解析
		id, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			log.Printf("⚠️ 行 %d: IDの解析に失敗しました: %v", i+2, err)
			continue
		}

		// 問題生成制限を解析
		limit, err := strconv.Atoi(record[4])
		if err != nil {
			log.Printf("⚠️ 行 %d: 問題生成制限の解析に失敗しました: %v", i+2, err)
			continue
		}

		// 問題生成カウントを解析
		generationCount, err := strconv.Atoi(record[5])
		if err != nil {
			log.Printf("⚠️ 行 %d: 問題生成カウントの解析に失敗しました: %v", i+2, err)
			continue
		}

		// 図形再生成制限を解析
		figureLimit, err := strconv.Atoi(record[6])
		if err != nil {
			log.Printf("⚠️ 行 %d: 図形再生成制限の解析に失敗しました: %v", i+2, err)
			continue
		}

		// 図形再生成カウントを解析
		figureCount, err := strconv.Atoi(record[7])
		if err != nil {
			log.Printf("⚠️ 行 %d: 図形再生成カウントの解析に失敗しました: %v", i+2, err)
			continue
		}

		// パスワードをハッシュ化
		passwordHash, err := utils.HashPassword(record[3])
		if err != nil {
			log.Printf("⚠️ 行 %d: パスワードのハッシュ化に失敗しました: %v", i+2, err)
			continue
		}

		user := &models.User{
			ID:                      id,
			SchoolCode:             record[1],
			Email:                  record[2],
			PasswordHash:           passwordHash,
			ProblemGenerationLimit: limit,
			ProblemGenerationCount: generationCount,
			FigureRegenerationLimit: figureLimit,
			FigureRegenerationCount: figureCount,
			Role:                   record[8],
			PreferredAPI:           record[9],
			PreferredModel:         record[10],
			CreatedAt:              now,
			UpdatedAt:              now,
		}

		users = append(users, user)
		log.Printf("📝 ユーザー追加: SchoolCode=%s, Email=%s, Role=%s, API=%s, Model=%s", 
			user.SchoolCode, user.Email, user.Role, user.PreferredAPI, user.PreferredModel)
	}

	return users, nil
}

func (r *memoryUserRepository) createDefaultUser() {
	// フォールバック用のデフォルトユーザーを作成
	now := time.Now()
	passwordHash, err := utils.HashPassword("password")
	if err != nil {
		log.Printf("❌ デフォルトユーザーのパスワードハッシュ化に失敗しました: %v", err)
		return
	}

	defaultUser := &models.User{
		ID:                      1,
		SchoolCode:             "00000",
		Email:                  "nutfes.script@gmail.com",
		PasswordHash:           passwordHash,
		ProblemGenerationLimit: 3,
		ProblemGenerationCount: 0,
		FigureRegenerationLimit: 2,
		FigureRegenerationCount: 0,
		Role:                   "teacher",
		PreferredAPI:           "claude",
		PreferredModel:         "claude-3-haiku",
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	r.users[defaultUser.SchoolCode] = defaultUser
	log.Printf("📝 デフォルトユーザーを作成しました: SchoolCode=%s", defaultUser.SchoolCode)
}

func (r *memoryUserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	for _, user := range r.users {
		if user.ID == id {
			return user, nil
		}
	}
	
	return nil, fmt.Errorf("user not found")
}

func (r *memoryUserRepository) GetBySchoolCode(ctx context.Context, schoolCode string) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	user, exists := r.users[schoolCode]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	
	return user, nil
}

func (r *memoryUserRepository) Create(ctx context.Context, user *models.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.users[user.SchoolCode]; exists {
		return fmt.Errorf("user with school code %s already exists", user.SchoolCode)
	}
	
	r.users[user.SchoolCode] = user
	return nil
}

func (r *memoryUserRepository) Update(ctx context.Context, user *models.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.users[user.SchoolCode]; !exists {
		return fmt.Errorf("user not found")
	}
	
	r.users[user.SchoolCode] = user
	return nil
}

func (r *memoryUserRepository) UpdateFigureRegenerationCount(userID int64, count int) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	for _, user := range r.users {
		if user.ID == userID {
			user.FigureRegenerationCount = count
			user.UpdatedAt = time.Now()
			return nil
		}
	}
	
	return fmt.Errorf("user not found")
}

// IncrementPreviewCount プレビュー回数をインクリメント
func (r *memoryUserRepository) IncrementPreviewCount(ctx context.Context, userID int64) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	for _, user := range r.users {
		if user.ID == userID {
			user.PreviewCount++
			user.UpdatedAt = time.Now()
			return nil
		}
	}
	
	return fmt.Errorf("user not found")
}
