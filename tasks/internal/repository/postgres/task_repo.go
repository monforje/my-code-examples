package postgresrepo

import (
	"context"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"tasks/internal/models/records"
	"tasks/internal/services"
)

type TaskRepo struct {
	*Repo
}

func NewTaskRepo(repo *Repo) *TaskRepo {
	return &TaskRepo{Repo: repo}
}

func (r *TaskRepo) Create(ctx context.Context, task *records.Task) error {
	_, err := r.Exec(ctx, `
		insert into tasks (id, task_name, title, description, specification_md_text, task_type, level, created_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
	`, task.ID, task.TaskName, task.Title, task.Description, task.SpecificationMDText, task.TaskType, task.Level, task.CreatedAt)
	return err
}

func (r *TaskRepo) GetByID(ctx context.Context, id uuid.UUID) (*records.Task, error) {
	task := new(records.Task)
	err := r.QueryRow(ctx, `
		select id, task_name, title, description, specification_md_text, task_type, level, created_at
		from tasks
		where id = $1
	`, id).Scan(
		&task.ID,
		&task.TaskName,
		&task.Title,
		&task.Description,
		&task.SpecificationMDText,
		&task.TaskType,
		&task.Level,
		&task.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (r *TaskRepo) GetByTaskName(ctx context.Context, taskName string) (*records.Task, error) {
	task := new(records.Task)
	err := r.QueryRow(ctx, `
		select id, task_name, title, description, specification_md_text, task_type, level, created_at
		from tasks
		where task_name = $1
	`, taskName).Scan(
		&task.ID,
		&task.TaskName,
		&task.Title,
		&task.Description,
		&task.SpecificationMDText,
		&task.TaskType,
		&task.Level,
		&task.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (r *TaskRepo) Update(ctx context.Context, task *records.Task) error {
	_, err := r.Exec(ctx, `
		update tasks
		set task_name = $2, title = $3, description = $4, specification_md_text = $5, task_type = $6, level = $7
		where id = $1
	`, task.ID, task.TaskName, task.Title, task.Description, task.SpecificationMDText, task.TaskType, task.Level)
	return err
}

func (r *TaskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.Exec(ctx, `
		delete from tasks
		where id = $1
	`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func (r *TaskRepo) List(ctx context.Context, limit int32, cursor *string, filters services.ListFilters) ([]records.TaskListItem, bool, error) {
	var query strings.Builder
	var args []any
	argIdx := 1
	var conditions []string

	if filters.Search != nil {
		search := strings.TrimSpace(*filters.Search)
		if search != "" {
			conditions = append(conditions, `(t.search_vector @@ websearch_to_tsquery('english', $`+strconv.Itoa(argIdx)+`)
				or t.title ilike '%' || $`+strconv.Itoa(argIdx)+` || '%'
				or t.description ilike '%' || $`+strconv.Itoa(argIdx)+` || '%')`)
			args = append(args, search)
			argIdx++
		}
	}

	if len(filters.Tags) > 0 {
		conditions = append(conditions, `exists (
			select 1 from task_tags tt
			join tags tg on tg.id = tt.tag_id
			where tt.task_id = t.id and (tt.tag_id::text = any($`+strconv.Itoa(argIdx)+`) or tg.name = any($`+strconv.Itoa(argIdx)+`))
		)`)
		args = append(args, filters.Tags)
		argIdx++
	}

	if len(filters.Languages) > 0 {
		conditions = append(conditions, `exists (
			select 1 from task_languages tl
			join languages l on l.id = tl.language_id
			where tl.task_id = t.id and (tl.language_id::text = any($`+strconv.Itoa(argIdx)+`) or l.name = any($`+strconv.Itoa(argIdx)+`))
		)`)
		args = append(args, filters.Languages)
		argIdx++
	}

	if filters.TaskType != nil {
		conditions = append(conditions, `t.task_type = $`+strconv.Itoa(argIdx))
		args = append(args, *filters.TaskType)
		argIdx++
	}

	if filters.Level != nil {
		conditions = append(conditions, `t.level = $`+strconv.Itoa(argIdx))
		args = append(args, *filters.Level)
		argIdx++
	}

	if cursor != nil {
		conditions = append(conditions, `(t.created_at, t.id) > (select c.created_at, c.id from tasks c where c.id = $`+strconv.Itoa(argIdx)+`)`)
		args = append(args, *cursor)
		argIdx++
	}

	query.WriteString(`select distinct t.id, t.task_name, t.title, t.description, t.task_type, t.level, t.created_at
		from tasks t`)

	if len(conditions) > 0 {
		query.WriteString(" where ")
		query.WriteString(strings.Join(conditions, " and "))
	}

	query.WriteString(" order by t.created_at asc, t.id asc")
	query.WriteString(" limit $" + strconv.Itoa(argIdx))
	args = append(args, limit+1)

	rows, err := r.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var items []records.TaskListItem
	for rows.Next() {
		var item records.TaskListItem
		if err := rows.Scan(&item.ID, &item.TaskName, &item.Title, &item.Description, &item.TaskType, &item.Level, &item.CreatedAt); err != nil {
			return nil, false, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasNextPage := int32(len(items)) > limit
	if hasNextPage {
		items = items[:int(limit)]
	}
	return items, hasNextPage, nil
}

func (r *TaskRepo) GetTagsByTaskID(ctx context.Context, taskID uuid.UUID) ([]records.Tag, error) {
	rows, err := r.Query(ctx, `
		select tg.id, tg.name
		from tags tg
		join task_tags tt on tt.tag_id = tg.id
		where tt.task_id = $1
		order by tg.name
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []records.Tag
	for rows.Next() {
		var tag records.Tag
		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (r *TaskRepo) SetTags(ctx context.Context, taskID uuid.UUID, tagIDs []uuid.UUID) error {
	_, err := r.Exec(ctx, `delete from task_tags where task_id = $1`, taskID)
	if err != nil {
		return err
	}

	for _, tagID := range tagIDs {
		_, err := r.Exec(ctx, `insert into task_tags (task_id, tag_id) values ($1, $2)`, taskID, tagID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *TaskRepo) GetLanguagesByTaskID(ctx context.Context, taskID uuid.UUID) ([]records.Language, error) {
	rows, err := r.Query(ctx, `
		select l.id, l.name
		from languages l
		join task_languages tl on tl.language_id = l.id
		where tl.task_id = $1
		order by l.name
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var langs []records.Language
	for rows.Next() {
		var lang records.Language
		if err := rows.Scan(&lang.ID, &lang.Name); err != nil {
			return nil, err
		}
		langs = append(langs, lang)
	}
	return langs, rows.Err()
}

func (r *TaskRepo) SetLanguages(ctx context.Context, taskID uuid.UUID, languageIDs []uuid.UUID) error {
	_, err := r.Exec(ctx, `delete from task_languages where task_id = $1`, taskID)
	if err != nil {
		return err
	}

	for _, langID := range languageIDs {
		_, err := r.Exec(ctx, `insert into task_languages (task_id, language_id) values ($1, $2)`, taskID, langID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *TaskRepo) GetTagsByIDs(ctx context.Context, ids []uuid.UUID) ([]records.Tag, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.Query(ctx, `select id, name from tags where id = any($1)`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []records.Tag
	for rows.Next() {
		var tag records.Tag
		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (r *TaskRepo) GetLanguagesByIDs(ctx context.Context, ids []uuid.UUID) ([]records.Language, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.Query(ctx, `select id, name from languages where id = any($1)`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var langs []records.Language
	for rows.Next() {
		var lang records.Language
		if err := rows.Scan(&lang.ID, &lang.Name); err != nil {
			return nil, err
		}
		langs = append(langs, lang)
	}
	return langs, rows.Err()
}

func (r *TaskRepo) ListTags(ctx context.Context) ([]records.Tag, error) {
	rows, err := r.Query(ctx, `select id, name from tags order by name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []records.Tag
	for rows.Next() {
		var tag records.Tag
		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (r *TaskRepo) ListLanguages(ctx context.Context) ([]records.Language, error) {
	rows, err := r.Query(ctx, `select id, name from languages order by name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var langs []records.Language
	for rows.Next() {
		var lang records.Language
		if err := rows.Scan(&lang.ID, &lang.Name); err != nil {
			return nil, err
		}
		langs = append(langs, lang)
	}
	return langs, rows.Err()
}
