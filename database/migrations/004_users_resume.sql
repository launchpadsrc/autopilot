-- +goose Up

alter table users add column resume jsonb;
alter table users add column resume_file bytea;

-- +goose Down

alter table users drop column resume, resume_file;
