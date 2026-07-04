import "@typespec/http";
import "@typespec/openapi";
using TypeSpec.Http;
using TypeSpec.OpenAPI;

namespace Tasks;

// ─── POST /tasks ──────────────────────────────────────────────────────────────
//
// curl -X POST http://localhost:8080/tasks \
//   -H "Authorization: Bearer <access_token>" \
//   -H "Content-Type: application/json" \
//   -d '{
//     "title": "Build auth service",
//     "description": "Implement JWT-based auth",
//     "specification_md_text": "# Specification\n\nImplement JWT-based auth...",
//     "task_type": "backend",
//     "level": "middle",
//     "tag_ids": ["tag-1", "tag-2"],
//     "language_ids": ["lang-1"]
//   }'
//

@route("/tasks")
@post
@operationId("tasks.create")
@useAuth(BearerAuth)
op createTask(@body body: CreateTaskRequest):
  | (Task & { @statusCode statusCode: 201; })
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// ─── GET /tasks ───────────────────────────────────────────────────────────────
//
// # первая страница (default limit=20)
// curl http://localhost:8080/tasks \
//   -H "Authorization: Bearer <access_token>"
//
// # поиск
// curl "http://localhost:8080/tasks?search=auth" \
//   -H "Authorization: Bearer <access_token>"
//
// # фильтр по тегам (OR)
// curl "http://localhost:8080/tasks?tags=tag-1&tags=tag-2" \
//   -H "Authorization: Bearer <access_token>"
//
// # фильтр по языкам (OR)
// curl "http://localhost:8080/tasks?languages=lang-1&languages=lang-2" \
//   -H "Authorization: Bearer <access_token>"
//
// # фильтр по типу и уровню
// curl "http://localhost:8080/tasks?task_type=backend&level=middle" \
//   -H "Authorization: Bearer <access_token>"
//

@route("/tasks")
@get
@operationId("tasks.list")
@useAuth(BearerAuth)
op listTasks(
  @query limit: int32 = 20,
  @query cursor?: string,
  @query search?: string,
  @query(#{explode: true}) tags?: string[],
  @query(#{explode: true}) languages?: string[],
  @query task_type?: TaskType,
  @query level?: Level,
):
  | TaskListResponse
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// ─── GET /tasks/{task_id} ─────────────────────────────────────────────────────
//
// curl http://localhost:8080/tasks/01JY4F4N5D6TQ7R8S9T0U1V2W3 \
//   -H "Authorization: Bearer <access_token>"
//

@route("/tasks/{task_id}")
@get
@operationId("tasks.get")
@useAuth(BearerAuth)
op getTask(@path task_id: string):
  | Task
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 404; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// ─── PATCH /tasks/{task_id} ───────────────────────────────────────────────────
//
// # обновить только title
// curl -X PATCH http://localhost:8080/tasks/01JY4F4N5D6TQ7R8S9T0U1V2W3 \
//   -H "Authorization: Bearer <access_token>" \
//   -H "Content-Type: application/json" \
//   -d '{"title": "Build users service"}'
//
// # обновить всё
// curl -X PATCH http://localhost:8080/tasks/01JY4F4N5D6TQ7R8S9T0U1V2W3 \
//   -H "Authorization: Bearer <access_token>" \
//   -H "Content-Type: application/json" \
//   -d '{
//     "title": "Build users service",
//     "description": "Updated description",
//     "specification_md_text": "# Updated spec",
//     "task_type": "frontend",
//     "level": "senior",
//     "tag_ids": ["tag-1"],
//     "language_ids": ["lang-1"]
//   }'
//

@route("/tasks/{task_id}")
@patch
@operationId("tasks.update")
@useAuth(BearerAuth)
op updateTask(@path task_id: string, @body body: UpdateTaskRequest):
  | Task
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 404; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// ─── DELETE /tasks/{task_id} ──────────────────────────────────────────────────
//
// curl -X DELETE http://localhost:8080/tasks/01JY4F4N5D6TQ7R8S9T0U1V2W3 \
//   -H "Authorization: Bearer <access_token>"
//

@route("/tasks/{task_id}")
@delete
@operationId("tasks.delete")
@useAuth(BearerAuth)
op deleteTask(@path task_id: string):
  | { @statusCode statusCode: 204; }
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 404; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// ─── GET /tags ────────────────────────────────────────────────────────────────
//
// curl http://localhost:8080/tags \
//   -H "Authorization: Bearer <access_token>"
//

@route("/tasks/tags")
@get
@operationId("tasks.tags.list")
@useAuth(BearerAuth)
op listTags():
  | Tag[]
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// ─── GET /tasks/languages ─────────────────────────────────────────────────────
//
// curl http://localhost:8080/tasks/languages \
//   -H "Authorization: Bearer <access_token>"
//

@route("/tasks/languages")
@get
@operationId("tasks.languages.list")
@useAuth(BearerAuth)
op listLanguages():
  | Language[]
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };
