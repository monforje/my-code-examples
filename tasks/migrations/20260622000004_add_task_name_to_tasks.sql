-- +goose Up

alter table tasks add column if not exists task_name text;

update tasks
set task_name = 'task-' || replace(id::text, '-', '')
where task_name is null or task_name = '';

update tasks
set task_name = 'pizza-api'
where id = 'c1000000-0000-0000-0000-000000000031';

alter table tasks alter column task_name set not null;

-- +goose StatementBegin
do $$
begin
    if not exists (
        select 1
        from pg_constraint
        where conname = 'tasks_task_name_key'
          and conrelid = 'tasks'::regclass
    ) then
        alter table tasks add constraint tasks_task_name_key unique (task_name);
    end if;
end $$;
-- +goose StatementEnd

-- +goose Down

alter table tasks drop constraint if exists tasks_task_name_key;
alter table tasks drop column if exists task_name;
