package services

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
)

type EmailService interface {
	SendEmail(to, subject, body string) error
}

type emailService struct {
	smtpHost     string
	smtpPort     string
	smtpFrom     string
	smtpPassword string
}

func NewEmailService() EmailService {
	return &emailService{
		smtpHost:     os.Getenv("SMTP_HOST"),
		smtpPort:     os.Getenv("SMTP_PORT"),
		smtpFrom:     os.Getenv("SMTP_FROM"),
		smtpPassword: os.Getenv("SMTP_PASSWORD"),
	}
}

func (s *emailService) SendEmail(to, subject, body string) error {
	// 設定値の詳細チェック
	if s.smtpFrom == "" {
		log.Printf("SMTP設定エラー: SMTP_FROM が設定されていません")
		return fmt.Errorf("SMTP_FROM が設定されていません")
	}
	if s.smtpPassword == "" || s.smtpPassword == "your-gmail-app-password" {
		log.Printf("SMTP設定エラー: SMTP_PASSWORD が設定されていないか、デフォルト値のままです")
		return fmt.Errorf("SMTP_PASSWORD が正しく設定されていません")
	}
	if s.smtpHost == "" {
		log.Printf("SMTP設定エラー: SMTP_HOST が設定されていません")
		return fmt.Errorf("SMTP_HOST が設定されていません")
	}
	if s.smtpPort == "" {
		log.Printf("SMTP設定エラー: SMTP_PORT が設定されていません")
		return fmt.Errorf("SMTP_PORT が設定されていません")
	}

	log.Printf("メール送信を開始します - To: %s, Subject: %s", to, subject)

	// メールメッセージの作成（正しいMIME形式）
	msg := []string{
		"From: " + s.smtpFrom,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}
	message := strings.Join(msg, "\r\n")

	// SMTP認証の設定
	auth := smtp.PlainAuth("", s.smtpFrom, s.smtpPassword, s.smtpHost)

	// TLS設定
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.smtpHost,
	}

	// SMTP接続の確立
	serverAddr := s.smtpHost + ":" + s.smtpPort
	log.Printf("SMTP接続を試行中: %s", serverAddr)

	conn, err := tls.Dial("tcp", serverAddr, tlsConfig)
	if err != nil {
		log.Printf("TLS接続エラー: %v", err)
		// TLSで失敗した場合、STARTTLSを試行
		return s.sendWithStartTLS(auth, serverAddr, s.smtpFrom, []string{to}, []byte(message))
	}

	client, err := smtp.NewClient(conn, s.smtpHost)
	if err != nil {
		log.Printf("SMTPクライアント作成エラー: %v", err)
		return fmt.Errorf("SMTPクライアントの作成に失敗しました: %w", err)
	}
	defer client.Quit()

	// 認証
	if err = client.Auth(auth); err != nil {
		log.Printf("SMTP認証エラー: %v", err)
		return fmt.Errorf("SMTP認証に失敗しました: %w", err)
	}

	// 送信者設定
	if err = client.Mail(s.smtpFrom); err != nil {
		log.Printf("送信者設定エラー: %v", err)
		return fmt.Errorf("送信者の設定に失敗しました: %w", err)
	}

	// 受信者設定
	if err = client.Rcpt(to); err != nil {
		log.Printf("受信者設定エラー: %v", err)
		return fmt.Errorf("受信者の設定に失敗しました: %w", err)
	}

	// メール本文送信
	writer, err := client.Data()
	if err != nil {
		log.Printf("データ送信開始エラー: %v", err)
		return fmt.Errorf("メール送信の開始に失敗しました: %w", err)
	}

	_, err = writer.Write([]byte(message))
	if err != nil {
		log.Printf("メール本文書き込みエラー: %v", err)
		return fmt.Errorf("メール本文の送信に失敗しました: %w", err)
	}

	err = writer.Close()
	if err != nil {
		log.Printf("メール送信完了エラー: %v", err)
		return fmt.Errorf("メール送信の完了に失敗しました: %w", err)
	}

	log.Printf("メールを正常に送信しました: %s", to)
	return nil
}

// STARTTLSを使用したメール送信のフォールバック
func (s *emailService) sendWithStartTLS(auth smtp.Auth, addr, from string, to []string, msg []byte) error {
	log.Printf("STARTTLSでの送信を試行中")
	
	err := smtp.SendMail(addr, auth, from, to, msg)
	if err != nil {
		log.Printf("STARTTLSメール送信エラー: %v", err)
		return fmt.Errorf("メール送信に失敗しました: %w", err)
	}
	
	log.Printf("STARTTLSでメールを正常に送信しました")
	return nil
}
