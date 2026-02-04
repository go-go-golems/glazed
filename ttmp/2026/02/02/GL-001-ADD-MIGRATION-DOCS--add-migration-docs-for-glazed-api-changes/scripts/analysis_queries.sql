-- API change analysis queries for git-diff-origin-main.sqlite

-- Counts of added/removed exported symbols.
select change_type, count(*) as count
from symbol_changes
group by change_type
order by change_type;

-- List removed exported symbols.
select path, package, kind, receiver, name
from symbol_changes
where change_type = 'removed'
order by path, name;

-- List added exported symbols.
select path, package, kind, receiver, name
from symbol_changes
where change_type = 'added'
order by path, name;

-- Top churn in pkg/ by raw diff line counts.
select path, status, additions, deletions, (additions + deletions) as total
from diff_files
where path like 'pkg/%'
order by total desc
limit 20;

-- Find diff hunks that mention ParsedLayers (signature shifts).
select path, hunk_index, hunk_text
from diff_hunks
where hunk_text like '%ParsedLayers%'
order by path, hunk_index;
