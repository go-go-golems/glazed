-- 06-most-read-files.sql
-- Files read most often (indicates confusion or iterative discovery)
SELECT
  regexp_extract(CAST(input AS VARCHAR), 'path:([^,}]+)', 1) AS file_path,
  COUNT(*) AS read_count
FROM sessions_base, UNNEST(tool_calls) AS t(tc)
WHERE json_extract(tc, '$.tool_name') = '"read"'
GROUP BY file_path
ORDER BY read_count DESC
LIMIT 30;
