-- +goose Up

create table if not exists tasks (
    id uuid primary key,
    task_name text not null unique,
    title text not null,
    description text not null default '',
    specification_md_text text not null,
    task_type text not null default 'backend'
        check (task_type in ('backend', 'frontend')),
    level text not null default 'middle'
        check (level in ('junior', 'middle', 'senior')),
    created_at timestamptz default now(),
    search_vector tsvector
        generated always as (
            to_tsvector('english', coalesce(title, '') || ' ' || coalesce(description, '') || ' ' || coalesce(task_name, ''))
        ) stored
);

create index idx_tasks_search on tasks using gin(search_vector);

-- +goose Down

drop index if exists idx_tasks_search;
drop table if exists tasks;
