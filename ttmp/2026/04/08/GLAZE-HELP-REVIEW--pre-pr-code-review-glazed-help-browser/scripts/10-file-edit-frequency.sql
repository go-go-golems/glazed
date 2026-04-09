-- 10-file-edit-frequency.sql
-- Which files were edited/written most often
SELECT
  regexp_extract(CAST(tc AS VARCHAR), 'path:(/home/[^ ]+)', 1) AS file_path,
  COUNT(*) AS op_count
FROM sessions_base, UNNEST(tool_calls) AS t(tc)
WHERE CAST(tc AS VARCHAR) LIKE '%operation_type:MODIFY%'
   OR CAST(tc AS VARCHAR) LIKE '%operation_type:CREATE%'
GROUP BY file_path
ORDER BY op_count DESC
LIMIT 50;
