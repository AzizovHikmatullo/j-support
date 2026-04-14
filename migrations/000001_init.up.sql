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
    content varchar(150) not null,
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

create index idx_activity_log_created_at on activity_log(created_at desc);
create index idx_activity_log_ticket_created on activity_log(ticket_id, created_at desc);

create index idx_categories_destination_enabled on categories(destination, enabled);

create index idx_bot_scenarios_category_active on bot_scenarios(category_id, is_active);

create index idx_bot_steps_scenario_id on bot_steps(scenario_id);
create index idx_bot_steps_parent_id on bot_steps(parent_id);
create index idx_bot_steps_scenario_parent_null on bot_steps(scenario_id) where parent_id is null;

create index idx_bot_sessions_ticket_id on bot_sessions(ticket_id);
create index idx_bot_sessions_last_activity on bot_sessions(last_activity_at);
create index idx_bot_sessions_scenario_id on bot_sessions(scenario_id);
create index idx_bot_sessions_current_step_id on bot_sessions(current_step_id);

create index idx_tickets_contact_created on tickets(contact_id, created_at desc);
create index idx_tickets_status_created on tickets(status, created_at desc);
create index idx_tickets_assigned_created on tickets(assigned_id, created_at desc);
create index idx_tickets_created_at on tickets(created_at desc);
create index idx_tickets_category_id on tickets(category_id);
create index idx_tickets_contact_id on tickets(contact_id);
create index idx_tickets_assigned_id on tickets(assigned_id);
create index idx_ticket_ratings_contact_id on ticket_ratings(contact_id);

create index idx_messages_ticket_created on messages(ticket_id, created_at);