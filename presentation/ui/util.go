package ui

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/swaggest/openapi-go"
	"go.wdy.de/nago/auth"
	"go.wdy.de/nago/logging"
	"log/slog"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

func writeJson(w http.ResponseWriter, r *http.Request, v any) {
	w.Header().Set("content-type", "application/json")
	enc := json.NewEncoder(w)
	if err := enc.Encode(v); err != nil {
		// TODO grab from request
		slog.Default().Error("failed to encode and write response", slog.Any("err", err))
	}
}

func setSummaryAndDescription(oc openapi.OperationContext, s string) {
	oc.SetDescription(s)
	sentences := strings.Split(s, ". ")
	if len(sentences) > 0 {
		summary := sentences[0]
		if len(summary) > 180 {
			summary = summary[:180] + "..."
		}
		oc.SetSummary(summary)
	}
}

// pathNames returns all path:"x" tagged public values from fields.
func pathNames(v any) []string {
	var res []string
	t := reflect.TypeOf(v)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		if s := f.Tag.Get("path"); s != "" {
			res = append(res, s)
		}
	}
	return res
}

func interpolatePathVariables[Params any](pattern string, r *http.Request) string {
	var params Params
	for _, pathVar := range pathNames(params) {
		pathVarValue := chi.URLParam(r, pathVar)
		pattern = strings.ReplaceAll(pattern, "{"+pathVar+"}", pathVarValue)
	}

	return pattern
}

var userT = reflect.TypeOf((*auth.User)(nil)).Elem()

func parseParams[P any](request *http.Request, authRequired bool) P {
	var params P
	t := reflect.TypeOf(params)
	v := reflect.ValueOf(&params).Elem()

	for _, pathVar := range pathNames(params) {
		pathVarValue := chi.URLParam(request, pathVar)
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if s := f.Tag.Get("path"); s == pathVar {
				switch f.Type.Kind() {
				case reflect.String:
					v.Field(i).SetString(pathVarValue)
				case reflect.Int:
					x, err := strconv.Atoi(pathVarValue)
					if err != nil {
						slog.Default().Error("cannot parse integer path variable", slog.Any("err", err))
					}

					v.Field(i).SetInt(int64(x))
				case reflect.Int64:
					x, err := strconv.ParseInt(pathVarValue, 10, 64)
					if err != nil {
						slog.Default().Error("cannot parse integer path variable", slog.Any("err", err))
					}

					v.Field(i).SetInt(x)

				case reflect.Uint64:
					x, err := strconv.ParseUint(pathVarValue, 10, 64)
					if err != nil {
						slog.Default().Error("cannot parse integer path variable", slog.Any("err", err))
					}

					v.Field(i).SetUint(x)
				default:
					logging.FromContext(request.Context()).Error(fmt.Sprintf("cannot parse path variable '%s' of type '%T' with value '%v'", pathVar, v.Field(i).Interface(), pathVarValue))
				}

			}

		}
	}

	user := auth.FromContext(request.Context())
	if authRequired && user == nil {
		logging.FromContext(request.Context()).Error("client did not provide a valid user, aborting request")
		panic(fmt.Errorf("security abort"))
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if s := f.Tag.Get("path"); s != "" {
			continue
		}

		if f.Type == userT {
			if user != nil {
				v.Field(i).Set(reflect.ValueOf(user))
				continue
			}
		}
		// not a path param
		logging.FromContext(request.Context()).Error(fmt.Sprintf("cannot parse or inject into unsupported field %T.%s of type %v", params, f.Name, f.Type.PkgPath()+"."+f.Type.Name()))
	}

	//TODO if params require a user and no user is available we must raise a not-authorized http response
	//TODO add query support
	//TODO add request/writer support to peak through the type system?

	return params
}
