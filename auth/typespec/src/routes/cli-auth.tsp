import "@typespec/http";
import "@typespec/openapi";
using TypeSpec.Http;
using TypeSpec.OpenAPI;

namespace Auth;

// ─── Device Authorization Grant ───────────────────────────────────────────────

// Запустить авторизацию CLI. Возвращает device_code и user_code для отображения в браузере.
@route("/auth/device/start")
@post
@operationId("auth.device.start")
op startDeviceAuth():
  | DeviceStartResponse
  | { @statusCode statusCode: 429; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// Подтвердить CLI-вход из браузера. Требует Bearer токен аутентифицированного пользователя.
@route("/auth/device/confirm")
@post
@operationId("auth.device.confirm")
@useAuth(BearerAuth)
op confirmDeviceAuth(@body body: DeviceConfirmRequest):
  | DeviceConfirmResponse
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 403; @body error: ErrorResponse; }
  | { @statusCode statusCode: 404; @body error: ErrorResponse; }
  | { @statusCode statusCode: 410; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 429; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// CLI опрашивает этот endpoint для получения токенов после подтверждения.
@route("/auth/device/token")
@post
@operationId("auth.device.token")
op pollDeviceToken(@body body: DeviceTokenRequest):
  | CliTokenResponse
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 404; @body error: ErrorResponse; }
  | { @statusCode statusCode: 428; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// ─── CLI Refresh ──────────────────────────────────────────────────────────────

// Обновить CLI access token через refresh token.
@route("/auth/cli/refresh")
@post
@operationId("auth.cli.refresh")
op refreshCliToken(@body body: CliRefreshRequest):
  | CliRefreshResponse
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 403; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };
