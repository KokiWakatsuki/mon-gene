-- opinion_profileカラムをproblemsテーブルに追加
ALTER TABLE problems
ADD COLUMN opinion_profile JSON COMMENT 'opinion.md基準のプロファイル（出題分野コード、コアスキル評価、問題構造評価、総合難易度スコア）' AFTER filters,
ADD COLUMN opinion_profile_v2 JSON COMMENT 'OpinionProfile Ver.2 (opinion_ver2.md基準) - 空間図形問題の詳細指標' AFTER opinion_profile,
ADD INDEX idx_opinion_domain ((CAST(JSON_EXTRACT(opinion_profile, '$.domain') AS UNSIGNED))),
ADD INDEX idx_opinion_difficulty ((CAST(JSON_EXTRACT(opinion_profile, '$.difficulty_score') AS UNSIGNED))),
ADD INDEX idx_opinion_v2_sub_problem_count ((CAST(JSON_EXTRACT(opinion_profile_v2, '$.sub_problem_count') AS UNSIGNED))),
ADD INDEX idx_opinion_v2_solution_steps ((CAST(JSON_EXTRACT(opinion_profile_v2, '$.solution_steps') AS UNSIGNED)));
