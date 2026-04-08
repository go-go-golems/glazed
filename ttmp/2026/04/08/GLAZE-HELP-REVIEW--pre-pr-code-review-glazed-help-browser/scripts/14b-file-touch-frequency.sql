-- 14b-file-touch-frequency.sql
-- Files touched most often across read/write/edit operations
-- tc is stored as VARCHAR, need to parse it
SELECT
  json_extract(tc, '$.tool_name') AS tool,
  json_extract(tc, '$.input.file_path') AS file_path,
  COUNT(*) AS count
FROM sessions_base, UNNEST(tool_calls) AS t(tc)
WHERE json_extract(tc, '$.tool_name') IN ('read', 'write', 'edit')
  AND json_extract(tc, '$.input.file_path') IS NOT NULL
GROUP BY tool, file_path
ORDER BY count DESC
LIMIT 50;
