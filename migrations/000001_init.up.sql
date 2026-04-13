create table categories (
    id serial primary key,
    name text not null,
    enabled boolean not null default true,
    destination text not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

create table contacts (
    id serial primary key,
    user_id text unique,
    external_id text unique,
    name text,
    phone text unique,
    created_at timestamp not null default now()
);

create table tickets (
    id uuid primary key,
    category_id integer not null references categories(id),
    contact_id int not null references contacts(id),
    assigned_id integer,
    status text not null,
    source text not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

create table ticket_ratings (
    id serial primary key ,
    ticket_id uuid not null unique references tickets(id) on delete cascade,
    contact_id int not null references contacts(id),
    score int not null check (score between 1 and 5),
    reason text
);

create table messages (
    id uuid primary key,
    ticket_id uuid not null references tickets(id),
    sender_id int not null,
    sender_type text not null,
    content text not null,
    created_at timestamp not null default now()
);

create table bot_scenarios (
    id serial primary key,
    category_id int not null references categories(id),
    is_active bool not null default true,
    created_at timestamp not null default now()
);

create table bot_steps (
    id serial primary key,
    scenario_id int not null references bot_scenarios(id) on delete cascade,
    parent_id int references  bot_steps(id) on delete cascade,
    condition text,
    question text not null,
    created_at timestamp not null default now()
);

create table bot_sessions (
    ticket_id uuid primary key references tickets(id) on delete cascade,
    scenario_id int not null references bot_scenarios(id),
    current_step_id int not null references bot_steps(id),
    created_at timestamp not null default now(),
    last_activity_at timestamp not null default now()
);

CREATE TABLE activity_log (
    id serial primary key,
    ticket_id uuid not null references tickets(id) on delete cascade,
    actor_id int not null,
    actor_type text not null,
    action text not null,
    payload jsonb,
    created_at timestamp not null default now()
);