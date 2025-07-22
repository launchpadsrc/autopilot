-- name: UpsertUserJob :exec
insert into user_jobs (
    user_id,
    job_id,
    feedback
) values (
    sqlc.arg(user_id),
    sqlc.arg(job_id),
    sqlc.arg(feedback)
) on conflict (user_id, job_id) do update set
    feedback = sqlc.arg(feedback),
    updated_at = now();