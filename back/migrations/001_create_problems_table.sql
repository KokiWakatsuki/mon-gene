-- ProblemsテーブルをCREATE
CREATE TABLE IF NOT EXISTS problems (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    subject VARCHAR(100) NOT NULL COMMENT '科目（数学、物理など）',
    prompt TEXT NOT NULL COMMENT '生成時のプロンプト',
    content TEXT NOT NULL COMMENT '問題文',
    solution TEXT COMMENT '解答',
    image_base64 LONGTEXT COMMENT '図（Base64エンコード）',
    conversation_history JSON COMMENT '5段階生成プロセスの会話履歴（図形再生成時に使用）',
    check_info JSON COMMENT 'チェック情報（問題文・解答・図のチェック状態、使用単元、年度、回数）',
    opinion_profile JSON COMMENT 'opinion.md基準のプロファイル（出題分野コード、コアスキル評価、問題構造評価、総合難易度スコア）',
    opinion_profile_v2 JSON COMMENT 'OpinionProfile Ver.2 (opinion_ver2.md基準) - 空間図形問題の詳細指標',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_subject (subject),
    INDEX idx_created_at (created_at),
    INDEX idx_opinion_domain ((CAST(JSON_EXTRACT(opinion_profile, '$.domain') AS UNSIGNED))),
    INDEX idx_opinion_difficulty ((CAST(JSON_EXTRACT(opinion_profile, '$.difficulty_score') AS UNSIGNED))),
    INDEX idx_opinion_v2_sub_problem_count ((CAST(JSON_EXTRACT(opinion_profile_v2, '$.sub_problem_count') AS UNSIGNED))),
    INDEX idx_opinion_v2_solution_steps ((CAST(JSON_EXTRACT(opinion_profile_v2, '$.solution_steps') AS UNSIGNED))),
    FULLTEXT INDEX idx_fulltext_search (content, solution, prompt, subject)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='問題テーブル';
