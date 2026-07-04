// Package build содержит значения, инъектируемые при сборке через ldflags -X.
// Значения по умолчанию рассчитаны на локальную разработку.
package build

var (
	Version     = "1.0.0"
	AuthAPIURL  = "http://codurity.ai/api/v1/auth"
	TasksAPIURL = "http://codurity.ai/api/v1"
	FrontendURL = "http://localhost:5173"
)
