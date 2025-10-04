package repositories

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mon-gene/back/internal/models"
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
	// 指定されたユーザーデータを追加
	now := time.Now()
	// パスワード "password" をbcryptでハッシュ化
	// bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost) の結果
	hashedPassword := "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi" // "password"のハッシュ
	
	user := &models.User{
		ID:           1,
		SchoolCode:   "00000",
		Email:        "nutfes.script@gmail.com",
		PasswordHash: hashedPassword,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	
	r.users[user.SchoolCode] = user
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
