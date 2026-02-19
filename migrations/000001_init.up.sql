create table categories (
    id serial primary key,
    name text not null,
    enabled boolean not null default true,
    destination text not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

create table tickets (
    id serial primary key,
    category_id integer not null references categories(id),
    creator_id integer not null,
    assigned_id integer,
    status text not null,
    subject text not null,
    source text not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);
create table messages (
    id bigserial primary key,
    ticket_id integer not null references tickets(id),
    sender_id integer not null,
    sender_type text not null,
    content text not null,
    created_at timestamp not null default now()
);