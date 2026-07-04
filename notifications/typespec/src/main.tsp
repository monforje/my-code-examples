import "@typespec/http";
import "@typespec/openapi3";

import "./models/common.tsp";
import "./models/ci_report.tsp";

import "./routes/ci_report.tsp";

using TypeSpec.Http;

@service(#{title: "Notification Service"})

namespace NotificationService;
