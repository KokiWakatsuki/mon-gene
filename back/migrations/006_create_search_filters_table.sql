-- 検索条件保存テーブル
CREATE TABLE IF NOT EXISTS search_filters (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL COMMENT '検索条件の名前',
    keyword VARCHAR(255) COMMENT 'キーワード検索',
    subject VARCHAR(50) COMMENT '科目',
    units JSON COMMENT '単元の配列',
    year VARCHAR(10) COMMENT '年度',
    exam_session VARCHAR(50) COMMENT '回数',
    is_checked BOOLEAN COMMENT 'チェック状態（true: チェック済み, false: 未チェック, null: すべて）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;