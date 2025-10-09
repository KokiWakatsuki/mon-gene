package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/mon-gene/back/internal/models"
	"github.com/mon-gene/back/internal/repositories"
	"github.com/mon-gene/back/internal/utils"
)

type AuthService interface {
	Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error)
	ForgotPassword(ctx context.Context, req models.ForgotPasswordRequest) (*models.ForgotPasswordResponse, error)
	ValidateToken(ctx context.Context, token string) (*models.User, error)
	Logout(ctx context.Context, token string) error
}

type authService struct {
	userRepo    repositories.UserRepository
	sessionRepo repositories.SessionRepository
	emailSvc    EmailService
}

func NewAuthService(
	userRepo repositories.UserRepository,
	sessionRepo repositories.SessionRepository,
	emailSvc EmailService,
) AuthService {
	return &authService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		emailSvc:    emailSvc,
	}
}

func (s *authService) Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
	// ユーザー取得
	user, err := s.userRepo.GetBySchoolCode(ctx, req.SchoolCode)
	if err != nil {
		return &models.LoginResponse{
			Success: false,
			Error:   "塾コードまたはパスワードが正しくありません",
		}, nil
	}

	// パスワード検証
	if !s.verifyPassword(req.Password, user.PasswordHash) {
		return &models.LoginResponse{
			Success: false,
			Error:   "塾コードまたはパスワードが正しくありません",
		}, nil
	}

	// トークン生成
	token, err := s.generateToken()
	if err != nil {
		return &models.LoginResponse{
			Success: false,
			Error:   "認証トークンの生成に失敗しました",
		}, nil
	}

	// セッション作成
	expiresAt := time.Now().Add(24 * time.Hour)
	if req.Remember {
		expiresAt = time.Now().Add(30 * 24 * time.Hour) // 30日間
	}

	session := &models.Session{
		ID:         token,
		UserID:     user.ID,
		SchoolCode: user.SchoolCode,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return &models.LoginResponse{
			Success: false,
			Error:   "セッションの作成に失敗しました",
		}, nil
	}

	return &models.LoginResponse{
		Success: true,
		Token:   token,
	}, nil
}

func (s *authService) ForgotPassword(ctx context.Context, req models.ForgotPasswordRequest) (*models.ForgotPasswordResponse, error) {
	// ユーザー取得
	user, err := s.userRepo.GetBySchoolCode(ctx, req.SchoolCode)
	if err != nil {
		return &models.ForgotPasswordResponse{
			Success: false,
			Error:   "指定された塾コードが見つかりません",
		}, nil
	}

	// 現在のパスワードを通知（本番環境では固定パスワード "password"）
	currentPassword := "password"

	// メール送信
	subject := "【Mongene】パスワードのお知らせ"
	body := fmt.Sprintf(`
こんにちは、

お忘れになったパスワードをお知らせいたします。

塾コード: %s
パスワード: %s

今後ともMongeneをよろしくお願いいたします。

Mongeneサポートチーム
`, user.SchoolCode, currentPassword)

	if err := s.emailSvc.SendEmail(user.Email, subject, body); err != nil {
		return &models.ForgotPasswordResponse{
			Success: false,
			Error:   fmt.Sprintf("メールの送信に失敗しました: %v", err),
		}, nil
	}

	return &models.ForgotPasswordResponse{
		Success: true,
		Message: "パスワードを記載したメールを送信しました",
	}, nil
}

func (s *authService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	// セッション取得
	session, err := s.sessionRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	// 有効期限チェック
	if time.Now().After(session.ExpiresAt) {
		s.sessionRepo.Delete(ctx, token) // 期限切れセッションを削除
		return nil, fmt.Errorf("token expired")
	}

	// セッションに保存されたSchoolCodeを使用してユーザーを取得
	user, err := s.userRepo.GetBySchoolCode(ctx, session.SchoolCode)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// セッションのUserIDと取得したユーザーのIDが一致するかチェック
	if user.ID != session.UserID {
		return nil, fmt.Errorf("user session mismatch")
	}

	return user, nil
}

func (s *authService) Logout(ctx context.Context, token string) error {
	return s.sessionRepo.Delete(ctx, token)
}

// generateToken generates a random token
func (s *authService) generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// verifyPassword verifies the password using bcrypt
func (s *authService) verifyPassword(password, hash string) bool {
	return utils.VerifyPassword(password, hash)
}

// hashPassword hashes a password using bcrypt
func (s *authService) hashPassword(password string) (string, error) {
	return utils.HashPassword(password)
}

// generateRandomPassword generates a random password
func (s *authService) generateRandomPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const passwordLength = 12
	
	bytes := make([]byte, passwordLength)
	for i := range bytes {
		randomIndex := make([]byte, 1)
		rand.Read(randomIndex)
		bytes[i] = charset[randomIndex[0]%byte(len(charset))]
	}
	return string(bytes)
}
