-- ProblemsテーブルをCREATE
CREATE TABLE IF NOT EXISTS problems (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    subject VARCHAR(100) NOT NULL COMMENT '科目（数学、物理など）',
    prompt TEXT NOT NULL COMMENT '生成時のプロンプト',
    content TEXT NOT NULL COMMENT '問題文',
    solution TEXT COMMENT '解答',
    image_base64 LONGTEXT COMMENT '図（Base64エンコード）',
    filters JSON COMMENT '生成パラメータ（フィルタ条件）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_subject (subject),
    INDEX idx_created_at (created_at),
    FULLTEXT INDEX idx_fulltext_search (content, solution, prompt, subject)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='問題テーブル';
