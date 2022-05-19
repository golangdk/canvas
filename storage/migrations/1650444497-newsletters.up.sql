create extension "pgcrypto";

create table newsletters (
    id uuid primary key default gen_random_uuid(),
    title text not null,
    body text not null,
    created timestamptz not null default now(),
    updated timestamptz not null default now()
);
