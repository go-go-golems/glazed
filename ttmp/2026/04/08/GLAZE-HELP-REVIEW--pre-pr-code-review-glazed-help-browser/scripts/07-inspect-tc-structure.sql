-- 07-inspect-tc-structure.sql
-- Check what's in tc for read operations
SELECT
  tc AS raw_tc
FROM sessions_base, UNNEST(tool_calls) AS t(tc)
WHERE json_extract(tc, '$.tool_name') = '"read"'
LIMIT 1;
