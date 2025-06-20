-- name: InsertUser :exec
insert into users (id) values ($1);

-- name: ResetUser :exec
update users set
    state = $2,
    state_dump = null,
    profile = null,
    resume = null,
    resume_file = null
where id = $1;

-- name: UserExists :one
select exists(select 1 from users where id = $1);

-- name: User :one
select * from users where id = $1;

-- name: UsersByState :many
select * from users where state = $1;

-- name: UpdateUserState :exec
update users set state = $2, state_dump = $3 where id = $1;

-- name: UpdateUserProfile :exec
update users set profile = $2 where id = $1;

-- name: UpdateUserResume :exec
update users set resume = $2, resume_file = $3 where id = $1;
