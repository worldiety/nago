package ui2

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/swaggest/openapi-go"
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

func parseParams[P any](request *http.Request) P {
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
				}
				if f.Type.Kind() == reflect.String {

				}

			}
		}
	}

	return params
}
