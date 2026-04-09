-- 16-file-touch-frequency.sql
-- Files touched most often across read/write/edit operations
SELECT
  tc.tool_name AS tool,
  tc.input.file_path AS file_path,
  COUNT(*) AS count
FROM sessions_base, UNNEST(tool_calls) AS t(tc)
WHERE tc.tool_name = 'read'
   OR tc.tool_name = 'write'
   OR tc.tool_name = 'edit'
GROUP BY tool, file_path
ORDER BY count DESC
LIMIT 50;
