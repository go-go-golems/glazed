-- 21-build-test-counts.sql
-- Summary counts of build/test cycles
SELECT
  CASE 
    WHEN CAST(json_extract(tc, '$.input.command') AS VARCHAR) LIKE '%go build%' THEN 'go-build'
    WHEN CAST(json_extract(tc, '$.input.command') AS VARCHAR) LIKE '%go test%' THEN 'go-test'
    WHEN CAST(json_extract(tc, '$.input.command') AS VARCHAR) LIKE '%pnpm build%' THEN 'pnpm-build'
    WHEN CAST(json_extract(tc, '$.input.command') AS VARCHAR) LIKE '%pnpm dev%' THEN 'pnpm-dev'
    WHEN CAST(json_extract(tc, '$.input.command') AS VARCHAR) LIKE '%npm run build%' THEN 'npm-build'
    WHEN CAST(json_extract(tc, '$.input.command') AS VARCHAR) LIKE '%npm run dev%' THEN 'npm-dev'
    WHEN CAST(json_extract(tc, '$.input.command') AS VARCHAR) LIKE '%vite%' THEN 'vite'
    ELSE 'other-build'
  END AS cmd_type,
  COUNT(*) AS count
FROM sessions_base, UNNEST(tool_calls) AS t(tc)
WHERE json_extract(tc, '$.tool_name') = '"bash"'
  AND (
    CAST(json_extract(tc, '$.input.command') AS VARCHAR) LIKE '%go build%'
    OR CAST(json_extract(tc, '$.input.command') AS VARCHAR) LIKE '%go test%'
    OR CAST(json_extract(tc, '$.input.command') AS VARCHAR) LIKE '%pnpm build%'
    OR CAST(json_extract(tc, '$.input.command') AS VARCHAR) LIKE '%pnpm dev%'
    OR CAST(json_extract(tc, '$.input.command') AS VARCHAR) LIKE '%npm run build%'
    OR CAST(json_extract(tc, '$.input.command') AS VARCHAR) LIKE '%npm run dev%'
    OR CAST(json_extract(tc, '$.input.command') AS VARCHAR) LIKE '%vite%'
  )
GROUP BY cmd_type
ORDER BY count DESC;
