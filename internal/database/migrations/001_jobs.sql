-- +goose Up

create table jobs
(
    created_at   timestamptz  not null default current_timestamp,
    id           varchar(255) primary key,
    source       varchar(255) not null,
    published_at timestamptz,
    link         text         not null,
    title        text         not null,
    description  text         not null,
    ai_company   varchar(255) not null default '',
    ai_role      varchar(255) not null default '',
    ai_seniority varchar(255) not null default '',
    ai_overview  text         not null,
    ai_hashtags  varchar(32)[] not null
);

-- +goose Down
