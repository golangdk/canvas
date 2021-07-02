create table newsletter_subscribers (
    email text primary key,
    token text not null,
    confirmed bool not null default false,
    active bool not null default true,
    created timestamp not null default now(),
    updated timestamp not null default now()
);
