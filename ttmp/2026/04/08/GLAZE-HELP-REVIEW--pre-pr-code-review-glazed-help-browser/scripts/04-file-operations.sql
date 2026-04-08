-- 04-file-operations.sql
-- What files did the session touch (read/write/edit)?
SELECT
  REPLACE(CAST(json_extract(tc, '$.tool_name') AS VARCHAR), '"', '') AS tool_name,
  REPLACE(CAST(json_extract(tc, '$.input.path') AS VARCHAR), '"', '') AS file_path,
  COUNT(*) AS count
FROM sessions_base, UNNEST(tool_calls) AS t(tc)
WHERE json_extract(tc, '$.tool_name') IN ('"read"', '"write"', '"edit"')
  AND json_extract(tc, '$.input.path') IS NOT NULL
GROUP BY tool_name, file_path
ORDER BY count DESC, tool_name
LIMIT 60;
