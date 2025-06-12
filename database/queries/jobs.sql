-- name: JobsExist :many
select id from jobs where id = any(sqlc.slice(ids));

-- name: InsertJob :exec
insert into jobs (
    id,
    source,
    published_at,
    link,
    title,
    description,
    ai_company,
    ai_role,
    ai_seniority,
    ai_overview,
    ai_hashtags
) values (
    sqlc.arg(id),
    sqlc.arg(source),
    sqlc.arg(published_at),
    sqlc.arg(link),
    sqlc.arg(title),
    sqlc.arg(description),
    sqlc.arg(ai_company),
    sqlc.arg(ai_role),
    sqlc.arg(ai_seniority),
    sqlc.arg(ai_overview),
    sqlc.arg(ai_hashtags)
);
