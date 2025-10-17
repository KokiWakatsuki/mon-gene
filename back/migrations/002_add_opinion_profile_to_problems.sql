-- opinion_profileカラムをproblemsテーブルに追加
ALTER TABLE problems 
ADD COLUMN opinion_profile JSON COMMENT 'opinion.md基準のプロファイル（出題分野コード、コアスキル評価、問題構造評価、総合難易度スコア）' AFTER filters,
ADD INDEX idx_opinion_domain ((CAST(JSON_EXTRACT(opinion_profile, '$.domain') AS UNSIGNED))),
ADD INDEX idx_opinion_difficulty ((CAST(JSON_EXTRACT(opinion_profile, '$.difficulty_score') AS UNSIGNED)));
