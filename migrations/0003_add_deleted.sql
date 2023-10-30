-- +goose Up

alter table urls
add column if not exists deleted boolean not null default false;
