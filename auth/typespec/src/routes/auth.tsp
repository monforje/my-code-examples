import "@typespec/http";
import "@typespec/openapi";
using TypeSpec.Http;
using TypeSpec.OpenAPI;

namespace Auth;

// ─── register ─────────────────────────────────────────────────────────────────

// Создать пользователя и отправить код подтверждения email
@route("/auth/register")
@post
@operationId("auth.register")
op register(@body body: RegisterRequest):
  | (RegisterResponse & { @statusCode statusCode: 201; })
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 409; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 429; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// Подтвердить регистрацию кодом из письма
@route("/auth/register/verify")
@post
@operationId("auth.register.verify")
op verifyRegistration(@body body: VerifyCodeRequest):
  | MessageResponse
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 404; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 429; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// Переотправить код подтверждения регистрации
@route("/auth/register/code/resend")
@post
@operationId("auth.register.code.resend")
op resendRegistrationCode(@body body: ResendCodeRequest):
  | MessageResponse
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 404; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 429; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// ─── login ────────────────────────────────────────────────────────────────────

// Войти по email и паролю
@route("/auth/login")
@post
@operationId("auth.login")
op login(@body body: LoginRequest):
  | TokenResponse
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 403; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 429; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// ─── refresh ──────────────────────────────────────────────────────────────────

// Обновить access token через refresh token из cookie
@route("/auth/refresh")
@post
@operationId("auth.refresh")
op refreshToken():
  | RefreshResponse
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 403; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// ─── logout ───────────────────────────────────────────────────────────────────

// Завершить сессию и инвалидировать refresh token
@route("/auth/logout")
@post
@operationId("auth.logout")
@useAuth(BearerAuth)
op logout():
  | { @statusCode statusCode: 204; }
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };
