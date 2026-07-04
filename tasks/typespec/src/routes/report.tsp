import "@typespec/http";
import "@typespec/openapi";
using TypeSpec.Http;
using TypeSpec.OpenAPI;

namespace Tasks;

// ─── POST /reports ────────────────────────────────────────────────────────────
//
// Внутренний сервисный вызов от notifications → tasks.
// Принимает нормализованный CI-отчёт для хранения и отображения в UI.
//
// curl -X POST http://localhost:8080/reports \
//   -H "Content-Type: application/json" \
//   -d '{
//     "uid": "alice/golden-pizza-api",
//     "commit": "abc123",
//     "status": "failed",
//     "created_at": "2026-06-24T12:00:00Z",
//     "summary": {
//       "status": "failed",
//       "message": "Application is not reachable on localhost:8080",
//       "root_cause": "APP_UNREACHABLE",
//       "passed": 4,
//       "failed": 3,
//       "blocked": 5,
//       "warnings": 1
//     },
//     "steps": [],
//     "lint_errors": [],
//     "warnings": [],
//     "raw_log_available": true
//   }'
//

@route("/reports")
@post
@operationId("reports.create")
op createReport(@body body: CreateReportRequest):
  | (Report & { @statusCode statusCode: 201; })
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// ─── GET /reports ────────────────────────────────────────────────────────────
//
// Список CI-отчётов текущего пользователя (identity из JWT).
// Если task_name указан — фильтрует по задаче, иначе — все отчёты.
// Курcорная пагинация (как у /tasks).
//
// curl "http://localhost:8080/reports?limit=20" \
//   -H "Authorization: Bearer <access_token>"
//

@route("/reports")
@get
@operationId("reports.list")
@useAuth(BearerAuth)
op listReports(
  @query task_name?: string,
  @query status?: ReportStatus,
  @query limit: int32 = 20,
  @query cursor?: string,
):
  | ReportListResponse
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// ─── GET /reports/{reportId} ──────────────────────────────────────────────────
//
// Один CI-отчёт по id (identity из JWT, проверка владельца по username).
// Используется фронтом для страницы детального отчёта и real-time опроса
// статуса pending → passed/failed без выборки всего списка.
//
// curl "http://localhost:8080/reports/550e8400-..." \
//   -H "Authorization: Bearer <access_token>"
//

@route("/reports/{reportId}")
@get
@operationId("reports.get")
@useAuth(BearerAuth)
op getReport(@path reportId: string):
  | Report
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 404; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };
