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

-- name: ScoredJobs :many
with ranked as (
    select
        j.id,
        j.source,
        j.published_at,
        j.link,
        j.title,
        j.description,
        j.ai_company,
        j.ai_role,
        j.ai_seniority,
        j.ai_overview,
        j.ai_hashtags,

        /* tech keywords overlap */
        cardinality(array(
            select unnest(j.ai_hashtags)
            intersect
            select unnest(sqlc.arg('hashtags')::text[])
        ))::numeric
        as tech_match,

        /* role keyword match */
        case
            when j.ai_role ilike any (sqlc.arg('role_patterns')::text[])
            then 1
            else 0
        end::numeric
        as role_match,

        /* simple seniority weight */
        case j.ai_seniority
            when 'Junior' then 1.0
            when 'Middle' then 0.7
            else               0.4
            end
        as seniority_boost
    from
        jobs as j
    left join user_jobs as uj
        on uj.job_id = j.id and uj.user_id = sqlc.arg('user_id')
    where
        uj.job_id is null
)
select
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
    ai_hashtags,
    (tech_match + role_match * 0.8 + seniority_boost)::double precision AS score
from
    ranked
order by
    score desc,
    published_at desc
limit
    sqlc.arg('limit');
