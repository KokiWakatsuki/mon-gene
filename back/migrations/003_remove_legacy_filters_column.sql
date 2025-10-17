-- 従来基準のfiltersカラムを削除し、opinion_profileのみを使用する
-- filtersカラム（従来の学年・難易度・公式数などの基準）を削除
ALTER TABLE problems DROP COLUMN filters;

-- filtersカラムに関連していたインデックスがあれば削除
-- （現在は特にfilters専用のインデックスはないので、この部分は不要）

-- コメント: opinion_profileカラムのみでVer.4.0基準の問題評価を行う
