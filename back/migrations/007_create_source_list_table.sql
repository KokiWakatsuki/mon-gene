-- 007_create_source_list_table.sql
-- ユーザーごとの出典リスト（年度・回数）を保存するテーブル

CREATE TABLE IF NOT EXISTS source_list (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    year VARCHAR(10) NOT NULL,
    exam_session VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_user_source (user_id, year, exam_session)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;