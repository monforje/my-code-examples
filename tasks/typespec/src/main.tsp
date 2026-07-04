import "@typespec/http";
import "@typespec/openapi3";

import "./models/common.tsp";
import "./models/tasks.tsp";
import "./models/git_tasks.tsp";
import "./models/report.tsp";

import "./routes/tasks.tsp";
import "./routes/git_tasks.tsp";
import "./routes/report.tsp";
using TypeSpec.Http;

@service(#{title: "Tasks Service"})

namespace Tasks;
