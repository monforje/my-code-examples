import "@typespec/http";
import "@typespec/openapi";
using TypeSpec.Http;
using TypeSpec.OpenAPI;

namespace NotificationService;

// Принять вебхук от task_runner (ci_started / ci_finished)
@route("/webhook")
@post
@operationId("webhook.receive")
op receive(@body body: WebhookRequest):
  | MessageResponse
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// Получить последний CI-отчёт по UID
@route("/reports/{uid}")
@get
@operationId("reports.get")
op getReport(@path uid: string):
  | CIReportResponse
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 404; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };
