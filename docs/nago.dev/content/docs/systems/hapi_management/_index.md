---
title: REST APIs (HAPI)
linkTitle: REST APIs
---

HAPI defines REST endpoints in Go and generates an OpenAPI 3.1 specification from them. You describe how a
request is read into a Go struct and how the response is written; HAPI registers the handler, documents the
types and maps errors to RFC 9457 problem responses. Together with [Token Management](../token_management/),
endpoints are protected with bearer tokens.

## Enable

```go
import cfghapi "go.wdy.de/nago/application/hapi/cfg"

api := std.Must(cfghapi.Enable(cfg)).API // *hapi.API
```

`cfghapi.Management` has the single field `API *hapi.API`. The title and version of the specification come
from `cfg.Name()` and the application version, the contact from the [theme settings](../theme_management/).

## Define an endpoint

```go
type HelloRequest struct {
	Name string
}

type HelloResponse struct {
	Greeting string `json:"greeting" doc:"The greeting."`
}

hapi.Get[HelloRequest](api, hapi.Operation{Path: "/api/v1/hello", Summary: "Say hello"}).
	Request(
		hapi.StrFromQuery(hapi.StrParam[HelloRequest]{Name: "name", IntoModel: func(dst *HelloRequest, value string) error {
			dst.Name = value
			return nil
		}}),
	).
	Response(hapi.ToJSON[HelloRequest, HelloResponse](func(in HelloRequest) (HelloResponse, error) {
		return HelloResponse{Greeting: "hello " + in.Name}, nil
	}))
```

- `hapi.Get`, `Post`, `Put` and `Delete` start an endpoint; `Response` registers it.
- Request options: `StrFromHeader`, `StrFromQuery`, `JSONFromBody`, `JSONFromFormField`, `FilesFromFormField`,
  `FromBinary`, `RawRequest` and `BearerAuth`.
- Response options: `ToJSON` and `ToBinary`.
- The struct tags `json`, `doc`, `required`, `example` and `supportingText` document the schema.
- `hapi.Doc(api, func(*oas.OpenAPI))` changes the generated specification directly.

## Documentation frontend

The specification is served at `/api/doc/spec.json`. Serve one of the bundled frontends to browse it at
`/api/doc/index.html`:

```go
cfg.Serve(stoplight.Dist()) // go.wdy.de/nago/pkg/stoplight, or pkg/swagger, pkg/redocly
```

Serve only one of them; with several, it is undefined which one is used.

## Permissions

HAPI declares no permissions and has no admin UI. Check the permissions in your handlers, for example with
the subject from `BearerAuth`.

## Related

- [Tutorial: REST](/docs/examples/tutorial-56-rest/)
- [Token Management](../token_management/)
