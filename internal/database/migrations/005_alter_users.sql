-- +goose Up
update users
set state_dump = coalesce(state_dump, '{}'),
    profile = coalesce(profile, '{}'),
    resume = coalesce(resume, '{}');

alter table users
    alter column state_dump set default '{}',
    alter column profile set default '{}',
    alter column resume set default '{}',
    alter column state_dump set not null,
    alter column profile set not null,
    alter column resume set not null;

-- +goose Down
alter table users
    alter column profile drop default,
    alter column profile drop not null,
    alter column state_dump drop default,
    alter column state_dump drop not null,
    alter column resume drop default,
    alter column resume drop not null;