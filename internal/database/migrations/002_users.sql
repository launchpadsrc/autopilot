-- +goose Up

create table users
(
    created_at timestamptz not null default current_timestamp,
    id         bigint      primary key,
    state      varchar(64) not null default '',
    state_dump jsonb,
    profile    jsonb
);

-- +goose Down
