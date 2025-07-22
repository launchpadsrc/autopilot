-- +goose Up

create type user_job_feedback as enum (
    'scored',
    'liked',
    'deleted'
);

create table user_jobs
(
    created_at timestamptz       not null default current_timestamp,
    updated_at timestamptz       not null default current_timestamp,
    user_id    bigint            not null references users (id),
    job_id     varchar(255)      not null references jobs (id),
    feedback   user_job_feedback not null default 'scored',

    primary key (user_id, job_id)
);

-- +goose Down
