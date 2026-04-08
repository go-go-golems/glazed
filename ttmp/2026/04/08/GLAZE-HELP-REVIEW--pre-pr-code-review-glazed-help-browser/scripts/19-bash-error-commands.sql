-- 19-bash-error-commands.sql
-- Bash commands that contained error-related keywords (failures, rebuilds)
SELECT
  json_extract(tc, '$.input.command') AS cmd,
  json_extract(tc, '$.timestamp') AS ts,
  json_extract(tc, '$.operation_type') AS op_type
FROM sessions_base, UNNEST(tool_calls) AS t(tc)
WHERE json_extract(tc, '$.tool_name') = '"bash"'
  AND (
    CAST(json_extract(tc, '$.input.command') AS VARCHAR) LIKE '%go build%'
    OR CAST(json_extract(tc, '$.input.command') AS VARCHAR) LIKE '%go test%'
    OR CAST(json_extract(tc, '$.input.command') AS VARCHAR) LIKE '%fail%'
    OR CAST(json_extract(tc, '$.input.command') AS VARCHAR) LIKE '%error%'
    OR CAST(json_extract(tc, '$.input.command') AS VARCHAR) LIKE '%npm run%'
    OR CAST(json_extract(tc, '$.input.command') AS VARCHAR) LIKE '%pnpm%'
    OR CAST(json_extract(tc, '$.input.command') AS VARCHAR) LIKE '%vite%'
  )
ORDER BY ts;
