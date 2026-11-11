// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package hapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"go.wdy.de/nago/pkg/oas/v31"
)

// echoSpec is an endpoint, which answers with its name and the values of the given path params.
func echoSpec(method, path, name string, params ...string) Spec {
	return Spec{
		Method:    method,
		Path:      path,
		Operation: &oas.Operation{Summary: name},
		Handler: func(w http.ResponseWriter, r *http.Request) {
			res := name
			for _, p := range params {
				res += "|" + p + "=" + r.PathValue(p)
			}
			_, _ = w.Write([]byte(res))
		},
	}
}

func mustReplace(t *testing.T, g *Group, specs ...Spec) {
	t.Helper()
	if err := g.Replace(nil, func(api *API) error {
		for _, s := range specs {
			if err := Handle(api, s); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func serve(g http.Handler, method, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	g.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
	return rec
}

func TestGroupDistributes(t *testing.T) {
	g := NewGroup("/api/areas/sales")
	mustReplace(t, g,
		echoSpec(http.MethodGet, "/orders/{id}", "get", "id"),
		echoSpec(http.MethodDelete, "/orders/{id}", "delete", "id"),
		echoSpec(http.MethodGet, "/orders/new", "new"),
		echoSpec(http.MethodPost, "/orders", "post"),
		echoSpec(http.MethodPut, "/orders", "put"),
		echoSpec(http.MethodGet, "/files/{path...}", "files", "path"),
		echoSpec(http.MethodGet, "/", "root"),
	)

	tests := []struct {
		name, method, target string
		status               int
		body                 string
		allow                string
	}{
		{"placeholder", "GET", "/api/areas/sales/orders/7", 200, "get|id=7", ""},
		{"other method at the same pattern", "DELETE", "/api/areas/sales/orders/7", 200, "delete|id=7", ""},
		{"literal is more specific than a placeholder", "GET", "/api/areas/sales/orders/new", 200, "new", ""},
		{"encoded slash is a part of the value", "GET", "/api/areas/sales/orders/a%2Fb", 200, "get|id=a/b", ""},
		{"rest of the path", "GET", "/api/areas/sales/files/a/b%2Fc", 200, "files|path=a/b/c", ""},
		{"root", "GET", "/api/areas/sales", 200, "root", ""},
		{"root with slash", "GET", "/api/areas/sales/", 200, "root", ""},
		{"trailing slash is another path", "GET", "/api/areas/sales/orders/7/", 404, "", ""},
		{"unknown path", "GET", "/api/areas/sales/customers", 404, "", ""},
		{"other prefix", "GET", "/api/areas/salesx/orders/7", 404, "", ""},
		{"no prefix", "GET", "/orders/7", 404, "", ""},
		{"other method", "PATCH", "/api/areas/sales/orders", 405, "", "POST, PUT"},
		{"HEAD is not GET", "HEAD", "/api/areas/sales/orders/7", 405, "", "DELETE, GET"},
		{"OPTIONS is not served", "OPTIONS", "/api/areas/sales/orders/new", 405, "", "DELETE, GET"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(g, tt.method, tt.target)
			if rec.Code != tt.status {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tt.status, rec.Body)
			}

			if tt.status == 200 {
				if rec.Body.String() != tt.body {
					t.Errorf("body = %q, want %q", rec.Body, tt.body)
				}
				return
			}

			if got := rec.Header().Get("Allow"); got != tt.allow {
				t.Errorf("Allow = %q, want %q", got, tt.allow)
			}

			var p ProblemDetails
			if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil && tt.method != "HEAD" {
				t.Fatalf("no problem details: %v: %s", err, rec.Body)
			}

			if tt.method != "HEAD" && (p.Status != tt.status || p.Code == "") {
				t.Errorf("problem details %+v", p)
			}
		})
	}
}

func TestGroupReplaceConflicts(t *testing.T) {
	tests := []struct {
		name  string
		specs []Spec
		ok    bool
	}{
		{"same pattern", []Spec{echoSpec("GET", "/x/{a}", "a"), echoSpec("GET", "/x/{b}", "b")}, false},
		{"common path, none more specific", []Spec{echoSpec("GET", "/x/{a}/y", "a"), echoSpec("GET", "/x/z/{b}", "b")}, false},
		{"same pattern, other methods", []Spec{echoSpec("GET", "/x/{a}", "a"), echoSpec("DELETE", "/x/{b}", "b")}, true},
		{"more specific", []Spec{echoSpec("GET", "/x/{a}", "a"), echoSpec("GET", "/x/y", "b")}, true},
		{"invalid path", []Spec{echoSpec("GET", "x", "a")}, false},
		{"placeholder within text", []Spec{echoSpec("GET", "/x{a}", "a")}, false},
		{"rest placeholder not at the end", []Spec{echoSpec("GET", "/{a...}/x", "a")}, false},
		{"described twice", []Spec{echoSpec("GET", "/x", "a"), echoSpec("GET", "/x", "b")}, false},
		{"unknown method", []Spec{echoSpec("FETCH", "/x", "a")}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGroup("/p")
			mustReplace(t, g, echoSpec("GET", "/old", "old"))

			err := g.Replace(nil, func(api *API) error {
				for _, s := range tt.specs {
					if err := Handle(api, s); err != nil {
						return err
					}
				}
				return nil
			})

			if (err == nil) != tt.ok {
				t.Fatalf("err = %v, want ok = %v", err, tt.ok)
			}

			// a failed replacement changes nothing, a successful one replaces the whole set
			old := serve(g, "GET", "/p/old")
			if tt.ok && old.Code != 404 {
				t.Errorf("old endpoint after a replacement: %d", old.Code)
			}

			if !tt.ok && (old.Code != 200 || g.Document().Paths["/old"] == nil) {
				t.Errorf("old endpoint after a failed replacement: %d", old.Code)
			}
		})
	}
}

func TestGroupDocument(t *testing.T) {
	g := NewGroup("/api/areas/sales")
	if rec := serve(http.HandlerFunc(g.ServeDocument), "GET", "/api/areas/sales/openapi.json"); rec.Code != 404 {
		t.Fatalf("document before the first replacement: %d", rec.Code)
	}

	order := &oas.Schema{Type: "object", Properties: map[string]*oas.Schema{"id": {Type: "string", Description: "the id"}}}
	spec := echoSpec("GET", "/orders/{id}", "get", "id")
	spec.Operation = &oas.Operation{
		Summary:    "Reads an order.",
		Tags:       []string{"orders"},
		Parameters: []oas.Parameter{{Name: "id", In: oas.LocationPath, Required: true, Schema: &oas.Schema{Type: "string"}}},
		Responses: oas.Responses{"200": {Content: map[string]oas.MediaType{
			"application/json": {Schema: &oas.Schema{Ref: "#/components/schemas/Order"}},
		}}},
	}
	spec.Schemas = map[string]*oas.Schema{"Order": order}

	type In struct{}
	type Out struct {
		Name string `json:"name"`
	}

	doc := &oas.OpenAPI{Info: oas.Info{Title: "sales"}, Tags: []oas.Tag{{Name: "orders", Description: "Orders of the area."}}}
	if err := g.Replace(doc, func(api *API) error {
		// the DSL and explicit endpoints share a group
		Get[In](api, Operation{Path: "/hello", Summary: "Says hello."}).
			Request().
			Response(ToJSON(func(in In) (Out, error) { return Out{Name: "hello"}, nil }))

		return Handle(api, spec)
	}); err != nil {
		t.Fatal(err)
	}

	if rec := serve(g, "GET", "/api/areas/sales/hello"); rec.Code != 200 || !strings.Contains(rec.Body.String(), `"hello"`) {
		t.Errorf("DSL endpoint in a group: %d %s", rec.Code, rec.Body)
	}

	rec := serve(http.HandlerFunc(g.ServeDocument), "GET", "/api/areas/sales/openapi.json")
	if rec.Code != 200 || !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("document: %d %s", rec.Code, rec.Header())
	}

	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}

	if got["openapi"] != oas.Version {
		t.Errorf("openapi = %v", got["openapi"])
	}

	servers, _ := got["servers"].([]any)
	if len(servers) != 1 || servers[0].(map[string]any)["url"] != "/api/areas/sales" {
		t.Errorf("servers = %v, want the prefix", got["servers"])
	}

	paths := got["paths"].(map[string]any)
	op := paths["/orders/{id}"].(map[string]any)["get"].(map[string]any)
	if op["summary"] != "Reads an order." || op["tags"].([]any)[0] != "orders" {
		t.Errorf("explicit operation = %v", op)
	}

	if paths["/hello"] == nil {
		t.Errorf("DSL operation is missing: %v", paths)
	}

	schemas := got["components"].(map[string]any)["schemas"].(map[string]any)
	if schemas["Order"].(map[string]any)["properties"].(map[string]any)["id"].(map[string]any)["description"] != "the id" {
		t.Errorf("explicit schema = %v", schemas["Order"])
	}

	// a replacement replaces the document as well
	mustReplace(t, g, echoSpec("GET", "/other", "other"))
	if d := g.Document(); d.Paths["/orders/{id}"] != nil || d.Paths["/other"] == nil || d.Components != nil {
		t.Errorf("document after the replacement: %+v", d)
	}
}

// TestGroupReplaceAtomic replaces the set while requests are served: every request is served by exactly one of the
// two sets, never by none or by a mix.
func TestGroupReplaceAtomic(t *testing.T) {
	g := NewGroup("")
	setA := []Spec{echoSpec("GET", "/a", "a"), echoSpec("GET", "/b", "a")}
	setB := []Spec{echoSpec("GET", "/a", "b"), echoSpec("GET", "/b", "b")}
	mustReplace(t, g, setA...)

	var wg sync.WaitGroup
	errs := make(chan error, 1)
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 500 {
				a, b := serve(g, "GET", "/a").Body.String(), serve(g, "GET", "/b").Body.String()
				if (a != "a" && a != "b") || (b != "a" && b != "b") {
					select {
					case errs <- errors.New("answer of no set: " + a + " " + b):
					default:
					}
					return
				}
			}
		}()
	}

	for i := range 200 {
		set := setA
		if i%2 == 0 {
			set = setB
		}
		mustReplace(t, g, set...)
	}

	wg.Wait()
	select {
	case err := <-errs:
		t.Fatal(err)
	default:
	}
}
