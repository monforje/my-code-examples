import "@typespec/http";
import "@typespec/openapi";
using TypeSpec.Http;
using TypeSpec.OpenAPI;

namespace Users;

// Получить свой профиль
@route("/profile/me")
@get
@operationId("users.profile.me.get")
@useAuth(BearerAuth)
op getProfile():
  | ProfileResponse
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// Обновить профиль / настройки профиля
@route("/profile/me/settings")
@patch
@operationId("users.profile.me.settings.update")
@useAuth(BearerAuth)
op updateProfileSettings(@body body: UpdateProfileSettingsRequest):
  | ProfileResponse
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// Загрузить или заменить аватар
@route("/profile/me/avatar")
@put
@operationId("users.profile.me.avatar.update")
@useAuth(BearerAuth)
op uploadAvatar(
  @header contentType: "multipart/form-data",
  @multipartBody body: {
    avatar: HttpPart<File>;
  },
):
  | UpdateAvatarResponse
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 413; @body error: ErrorResponse; }
  | { @statusCode statusCode: 422; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };

// Удалить аватар
@route("/profile/me/avatar")
@delete
@operationId("users.profile.me.avatar.delete")
@useAuth(BearerAuth)
op deleteAvatar():
  | DeleteAvatarResponse
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };
