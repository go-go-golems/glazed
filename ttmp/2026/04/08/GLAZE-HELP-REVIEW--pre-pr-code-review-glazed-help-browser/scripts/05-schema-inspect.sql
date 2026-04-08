-- 05-schema-inspect.sql
-- Inspect the schema and sample tool_calls structure
SELECT json_extract(tc, '$.tool_name') AS tool_name,
       json_extract(tc, '$.input') AS input_sample
FROM sessions_base, UNNEST(tool_calls) AS t(tc)
WHERE json_extract(tc, '$.tool_name') = '"read"'
LIMIT 3;
