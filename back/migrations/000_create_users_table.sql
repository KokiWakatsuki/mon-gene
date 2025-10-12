-- Usersテーブルを作成
CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    school_code VARCHAR(10) NOT NULL UNIQUE COMMENT '学校コード',
    email VARCHAR(255) NOT NULL COMMENT 'メールアドレス',
    password_hash VARCHAR(255) NOT NULL COMMENT 'ハッシュ化されたパスワード',
    problem_generation_limit INT NOT NULL DEFAULT 10 COMMENT '問題生成制限回数（-1 = 制限なし, 0以上 = 制限回数）',
    problem_generation_count INT NOT NULL DEFAULT 0 COMMENT '現在の問題生成回数',
    figure_regeneration_limit INT NOT NULL DEFAULT 2 COMMENT '図形再生成制限回数（-1 = 制限なし, 0以上 = 制限回数）',
    figure_regeneration_count INT NOT NULL DEFAULT 0 COMMENT '現在の図形再生成回数',
    role VARCHAR(50) NOT NULL DEFAULT 'teacher' COMMENT 'ユーザーロール（admin, developer, teacher）',
    preferred_api VARCHAR(50) NOT NULL DEFAULT 'claude' COMMENT '優先API（chatgpt, claude, gemini）',
    preferred_model VARCHAR(100) NOT NULL DEFAULT 'claude-3-haiku' COMMENT '優先モデル名',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_school_code (school_code),
    INDEX idx_email (email),
    INDEX idx_role (role),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='ユーザーテーブル';
