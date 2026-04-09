-- 15-inspect-types.sql
-- Check types of tool_calls elements
SELECT 
  typeof(tool_calls) AS arr_type,
  length(tool_calls) AS arr_len,
  typeof(tool_calls[5]) AS elem_type,
  SUBSTR(CAST(tool_calls[5] AS VARCHAR), 1, 100) AS elem_sample
FROM sessions_base
LIMIT 1;
