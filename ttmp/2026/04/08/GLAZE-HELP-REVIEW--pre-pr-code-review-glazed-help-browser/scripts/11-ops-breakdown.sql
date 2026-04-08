-- 11-ops-breakdown.sql
-- Break down by operation_type  
SELECT
  regexp_extract(CAST(tc AS VARCHAR), 'operation_type:(\w+)', 1) AS op_type,
  COUNT(*) AS count
FROM sessions_base, UNNEST(tool_calls) AS t(tc)
GROUP BY op_type
ORDER BY count DESC;
