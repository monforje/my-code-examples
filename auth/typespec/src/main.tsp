import "@typespec/http";
import "@typespec/openapi3";

import "./models/common.tsp";
import "./models/auth.tsp";
import "./models/password.tsp";
import "./models/me.tsp";
import "./models/cli-auth.tsp";

import "./routes/auth.tsp";
import "./routes/password-forgot.tsp";
import "./routes/password-change.tsp";
import "./routes/me.tsp";
import "./routes/cli-auth.tsp";

using TypeSpec.Http;

@service(#{title: "Auth Service"})

namespace Auth;
