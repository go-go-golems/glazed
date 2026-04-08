-- 09-check-tc-format.sql
-- Check how tool_calls is stored 
SELECT 
  typeof(tool_calls) AS tc_type,
  typeof(tool_calls[1]) AS elem_type,
  tool_calls[1] AS first_elem
FROM sessions_base
LIMIT 1;
