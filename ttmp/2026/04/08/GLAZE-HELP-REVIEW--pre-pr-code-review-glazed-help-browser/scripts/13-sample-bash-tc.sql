-- 13-sample-bash-tc.sql
-- Sample bash tool calls - check what tool_name looks like
SELECT 
  SUBSTR(CAST(tc AS VARCHAR), 1, 300) AS sample
FROM sessions_base, UNNEST(tool_calls) AS t(tc)
WHERE CAST(tc AS VARCHAR) LIKE '%bash%'
LIMIT 3;
