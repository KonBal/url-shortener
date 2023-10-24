-- +goose Up

create table if not exists urls (
	id serial primary key,
	short_url varchar not null,
	original_url varchar not null unique
);
