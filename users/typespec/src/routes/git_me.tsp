import "@typespec/http";
import "@typespec/openapi";
using TypeSpec.Http;
using TypeSpec.OpenAPI;

namespace Users;

// Внутренний endpoint. Нельзя открывать наружу.
@route("/git-user/me")
@get
@operationId("users.git.me.get")
@useAuth(ApiKeyAuth<ApiKeyLocation.header, "X-Service-Token">)
op getGitMe(@body body: GitMeRequest):
  | GitMeResponse
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 404; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };
