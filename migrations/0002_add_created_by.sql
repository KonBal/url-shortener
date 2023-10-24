-- +goose Up

alter table urls
add column if not exists created_by varchar;
