import "@typespec/http";
import "@typespec/openapi3";

import "./models/common.tsp";
import "./models/me.tsp";
import "./models/git_me.tsp";

import "./routes/me.tsp";
import "./routes/git_me.tsp";

using TypeSpec.Http;

@service(#{title: "Users Service"})

namespace Users;
