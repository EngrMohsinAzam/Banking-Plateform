package api

import _ "embed"

// OpenAPISpec is the embedded OpenAPI 3 contract for the banking HTTP API.
//
//go:embed openapi.yaml
var OpenAPISpec []byte
