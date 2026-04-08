-- 12-sample-bash-tc.sql
-- Sample bash tool calls to understand structure
SELECT 
  SUBSTR(CAST(tc AS VARCHAR), 1, 200) AS sample
FROM sessions_base, UNNEST(tool_calls) AS t(tc)
WHERE CAST(tc AS VARCHAR) LIKE '%tool_name:bash%'
LIMIT 3;
