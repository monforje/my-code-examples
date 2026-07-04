import "@typespec/http";
import "@typespec/openapi";
using TypeSpec.Http;
using TypeSpec.OpenAPI;

namespace Auth;

// Запросить код восстановления пароля на email
@route("/auth/password/forgot")
@post
@operationId("auth.password.forgot")
op forgotPassword(@body body: ForgotPasswordRequest):
  | MessageResponse
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 429; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// Проверить код восстановления и получить reset token
@route("/auth/password/forgot/verify")
@post
@operationId("auth.password.forgot.verify")
op verifyForgotPassword(@body body: ForgotVerifyRequest):
  | ResetTokenResponse
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 429; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// Переотправить код восстановления пароля
@route("/auth/password/forgot/code/resend")
@post
@operationId("auth.password.forgot.code.resend")
op resendForgotPasswordCode(@body body: ResendCodeRequest):
  | MessageResponse
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 429; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// Установить новый пароль. Инвалидирует все refresh tokens пользователя.
@route("/auth/password/reset")
@post
@operationId("auth.password.reset")
op resetPassword(@body body: ResetPasswordRequest):
  | MessageResponse
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };
