-- 08-most-read-files.sql
-- Files read most often
SELECT
  regexp_extract(tc, 'path:([^ ]+)', 1) AS file_path,
  COUNT(*) AS read_count
FROM sessions_base, UNNEST(tool_calls) AS t(tc)
WHERE regexp_extract(tc, 'operation_type:(\w+)', 1) = 'READ'
GROUP BY file_path
ORDER BY read_count DESC
LIMIT 40;
