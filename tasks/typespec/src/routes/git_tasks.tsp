import "@typespec/http";
import "@typespec/openapi";
using TypeSpec.Http;
using TypeSpec.OpenAPI;

namespace Tasks;

@route("/tasks/{task_name}/git")
@post
@operationId("tasks.git.create")
@useAuth(BearerAuth)
op createGitTask(@path task_name: string, @body body: GitTaskRequest):
  | GitTaskResponse
  | { @statusCode statusCode: 400; @body error: ErrorResponse; }
  | { @statusCode statusCode: 401; @body error: ErrorResponse; }
  | { @statusCode statusCode: 404; @body error: ErrorResponse; }
  | { @statusCode statusCode: 409; @body error: ErrorResponse; }
  | { @statusCode statusCode: 500; @body error: ErrorResponse; };
