import "@typespec/http";
import "@typespec/openapi";
using TypeSpec.Http;
using TypeSpec.OpenAPI;

namespace Auth;

// Начать смену пароля для авторизованного пользователя. Проверяет текущий пароль и отправляет код подтверждения.
@route("/auth/password/change")
@post
@operationId("auth.password.change")
@useAuth(BearerAuth)
op changePassword(@body body: ChangePasswordRequest):
  | MessageResponse
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 403; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 429; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// Подтвердить код смены пароля и получить временный токен смены пароля.
@route("/auth/password/change/verify")
@post
@operationId("auth.password.change.verify")
@useAuth(BearerAuth)
op verifyChangePassword(@body body: CodeVerifyRequest):
  | ChangePasswordTokenResponse
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 403; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 429; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// Завершить смену пароля по временному токену. Инвалидирует refresh-сессии пользователя.
@route("/auth/password/change/complete")
@post
@operationId("auth.password.change.complete")
@useAuth(BearerAuth)
op completeChangePassword(@body body: CompletePasswordChangeRequest):
  | MessageResponse
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// Переотправить код подтверждения смены пароля
@route("/auth/password/change/code/resend")
@post
@operationId("auth.password.change.code.resend")
@useAuth(BearerAuth)
op resendChangePasswordCode():
  | MessageResponse
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 403; @body error: ErrorResponse; }
  | { @statusCode statusCode: 404; @body error: ErrorResponse; }
  | { @statusCode statusCode: 429; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };
