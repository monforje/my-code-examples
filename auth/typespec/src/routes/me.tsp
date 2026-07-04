import "@typespec/http";
import "@typespec/openapi";
using TypeSpec.Http;
using TypeSpec.OpenAPI;

namespace Auth;

// Получить текущую identity (кто я с точки зрения auth)
@route("/auth/me")
@get
@operationId("auth.me.get")
@useAuth(BearerAuth)
op getMe():
  | Identity
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 403; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// Начать удаление аккаунта. Отправляет код подтверждения.
@route("/auth/me")
@delete
@operationId("auth.me.delete")
@useAuth(BearerAuth)
op deleteMe(@body body: DeleteAccountRequest):
  | MessageResponse
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 403; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// Подтвердить удаление аккаунта. Удаляет аккаунт и инвалидирует все токены.
@route("/auth/me/delete/verify")
@post
@operationId("auth.me.delete.verify")
@useAuth(BearerAuth)
op verifyDeleteAccount(@body body: CodeVerifyRequest):
  | { @statusCode statusCode: 204; }
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 404; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 429; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// Переотправить код подтверждения удаления аккаунта
@route("/auth/me/delete/code/resend")
@post
@operationId("auth.me.delete.code.resend")
@useAuth(BearerAuth)
op resendDeleteAccountCode():
  | MessageResponse
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 403; @body error: ErrorResponse; }
  | { @statusCode statusCode: 404; @body error: ErrorResponse; }
  | { @statusCode statusCode: 429; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// ─── Email Change ─────────────────────────────────────────────────────────────

// Начать смену email. Проверяет пароль и отправляет код на ТЕКУЩИЙ email.
@route("/auth/me/email/change")
@post
@operationId("auth.me.email.change")
@useAuth(BearerAuth)
op changeEmail(@body body: EmailChangeRequest):
  | MessageResponse
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 403; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// Подтвердить код с текущего email и получить identity_token.
@route("/auth/me/email/change/verify")
@post
@operationId("auth.me.email.change.verify")
@useAuth(BearerAuth)
op verifyChangeEmail(@body body: CodeVerifyRequest):
  | IdentityTokenResponse
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 429; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// Указать новый email и получить код подтверждения на него.
@route("/auth/me/email/change/confirm")
@post
@operationId("auth.me.email.change.confirm")
@useAuth(BearerAuth)
op confirmChangeEmail(@body body: ConfirmEmailChangeRequest):
  | MessageResponse
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 409; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 429; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// Завершить смену email по коду с нового email. Отзывает остальные сессии.
@route("/auth/me/email/change/complete")
@post
@operationId("auth.me.email.change.complete")
@useAuth(BearerAuth)
op completeChangeEmail(@body body: CompleteEmailChangeRequest):
  | MessageResponse
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 403; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 429; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// Переотправить код подтверждения смены email
@route("/auth/me/email/change/code/resend")
@post
@operationId("auth.me.email.change.code.resend")
@useAuth(BearerAuth)
op resendChangeEmailCode(@query step: "current" | "new"):
  | MessageResponse
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 403; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 429; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };
