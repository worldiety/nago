package ui

import (
	"fmt"
	"github.com/swaggest/openapi-go/openapi3"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"reflect"
	"runtime/debug"
)

type Form[FormType, PageParams any] struct {
	ID          ComponentID
	Description string
	Init        func(PageParams) FormType
	Load        func(FormType, PageParams) FormType
	Submit      FormAction[FormType, PageParams]
	Delete      FormAction[FormType, PageParams]
	MaxMemory   int64
}

func (f Form[FormType, PageParams]) ComponentID() ComponentID {
	return f.ID
}

func (f Form[FormType, PageParams]) configure(parentSlug string, r router) {
	pattern := filepath.Join(parentSlug, string(f.ID))
	metaForm := formResponse{Type: "Form"}

	if f.Load != nil {
		metaForm.Links.Load = Link(filepath.Join(pattern, "load"))
		r.MethodFunc(http.MethodGet, string(metaForm.Links.Load), func(writer http.ResponseWriter, request *http.Request) {
			loadResp := formLoadResponse{}
			params := parseParams[PageParams](request)
			var zeroForm FormType
			if f.Init != nil {
				zeroForm = f.Init(params)
			}

			if f.Load != nil {
				zeroForm = f.Load(zeroForm, params)
			}
			loadResp.Fields = collectFields(zeroForm)
			writeJson(writer, request, loadResp)
		})
	} else {
		slog.Default().Warn(fmt.Sprintf("the form '%s' has no Load func and will not work properly", f.ID))
	}

	if f.Submit.Receive != nil || f.Delete.Receive != nil {
		if f.Submit.Receive != nil {
			metaForm.SubmitText = f.Submit.Title
			metaForm.Links.Submit = Link(filepath.Join(pattern, "submit"))

		}

		if f.Delete.Receive != nil {

			metaForm.DeleteText = f.Delete.Title
			metaForm.Links.Delete = Link(filepath.Join(pattern, "delete"))
		}
		r.MethodFunc(http.MethodPost, string(metaForm.Links.Submit), func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println(r)
					debug.PrintStack()
				}
			}()
			slog.Default().Info("hello upload")
			maxMem := f.MaxMemory
			if maxMem <= 0 {
				maxMem = 1024 * 1024 * 64
			}
			if err := r.ParseMultipartForm(maxMem); err != nil {
				slog.Default().Error("failed to parse multipart form", slog.Any("err", err)) // TODO
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if r.MultipartForm == nil {
				slog.Default().Error("not multipart but was expected") // TODO
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			for s, strings := range r.MultipartForm.Value {
				fmt.Printf("%s=>%v\n", s, strings)
			}

			for s, headers := range r.MultipartForm.File {
				fmt.Printf("%s=>%v\n", s, headers)
			}

			var zeroForm FormType
			pageParams := parseParams[PageParams](r)
			if f.Init != nil {
				zeroForm = f.Init(pageParams)
			}

			if err := unmarshalForm(&zeroForm, r.MultipartForm); err != nil {
				slog.Default().Error("failed to unmarshal form", slog.Any("err", err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var formToFix FormType
			var action Action
			switch r.MultipartForm.Value["_action"][0] {
			case "delete":
				formToFix, action = f.Delete.Receive(zeroForm, pageParams)
			case "update":
				formToFix, action = f.Submit.Receive(zeroForm, pageParams)
			default:
				panic(fmt.Errorf("invalid action: %v", r.MultipartForm.Value["_action"]))
			}

			if action == nil {
				submitResp := formSubmitResponse{
					Type: "FormValidationError",
				}
				submitResp.Fields = collectFields(formToFix)
				writeJson(w, r, submitResp)
				return
			}

			writeJson(w, r, action)

		})
	}

	r.MethodFunc(http.MethodGet, pattern, func(writer http.ResponseWriter, request *http.Request) {
		tmp := metaForm
		tmp.Links.Load = Link(interpolatePathVariables[PageParams](string(tmp.Links.Load), request))
		tmp.Links.Submit = Link(interpolatePathVariables[PageParams](string(tmp.Links.Submit), request))
		tmp.Links.Delete = Link(interpolatePathVariables[PageParams](string(tmp.Links.Delete), request))
		writeJson(writer, request, tmp)
	})

}

func (f Form[FormType, PageParams]) renderOpenAPI(p PageParams, tag string, parentSlug string, r *openapi3.Reflector) {
	pattern := filepath.Join(parentSlug, string(f.ID))
	oc := must2(r.NewOperationContext(http.MethodGet, pattern))
	oc.AddReqStructure(p)
	oc.AddRespStructure(formResponse{})
	oc.SetTags(tag)
	setSummaryAndDescription(oc, f.Description)
	must(r.AddOperation(oc))

	if f.Load != nil {
		oc := must2(r.NewOperationContext(http.MethodGet, filepath.Join(pattern, "load")))
		oc.AddReqStructure(p)
		oc.AddRespStructure(formLoadResponse{})
		oc.SetTags(tag)
		must(r.AddOperation(oc))
	}
}

type formSubmitResponse struct {
	Type   string      `json:"type"`
	Fields []inputType `json:"fields"`
}

type formResponse struct {
	Type       string `json:"type"`
	SubmitText string `json:"submitText"`
	DeleteText string `json:"deleteText"`
	Links      struct {
		Load   Link `json:"load,omitempty"`
		Submit Link `json:"submit,omitempty"`
		Delete Link `json:"delete,omitempty"`
	} `json:"links"`
}

type formLoadResponse struct {
	Fields []inputType `json:"fields"`
}

type inputType struct {
	Type           string            `json:"type"`
	ID             string            `json:"id"`
	Label          string            `json:"label"`
	Value          string            `json:"value,omitempty"`
	Hint           string            `json:"hint"`
	Error          string            `json:"error"`
	LayoutHints    layoutHints       `json:"layoutHints"`
	Disabled       bool              `json:"disabled"`
	FileMultiple   bool              `json:"fileMultiple,omitempty"`
	FileAccept     string            `json:"fileAccept,omitempty"`
	SelectMultiple bool              `json:"selectMultiple"`
	SelectItems    []inputSelectItem `json:"selectItems"`
	SelectValues   []string          `json:"selectValues"`
}

type inputSelectItem struct {
	ID      string `json:"value"`
	Caption string `json:"title"`
}

type layoutHints struct {
	CSS css `json:"css"`
}

type css struct {
	Class string `json:"class"`
}

func collectFields(f any) []inputType {
	var res []inputType
	t := reflect.TypeOf(f)
	v := reflect.ValueOf(f)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		if input, ok := v.Field(i).Interface().(InputField); ok {
			ip := input.intoInput()
			ip.ID = field.Name
			if c := field.Tag.Get("class"); c != "" {
				ip.LayoutHints.CSS.Class = c
			}
			res = append(res, ip)
		}

	}

	return res
}

func unmarshalForm(dst any, form *multipart.Form) error {
	t := reflect.TypeOf(dst).Elem()
	v := reflect.ValueOf(dst).Elem()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		if input, ok := v.Field(i).Interface().(InputField); ok {
			if values, ok := form.Value[field.Name]; ok && len(values) > 0 {
				nVal := input.setValue(values[0])
				v.Field(i).Set(reflect.ValueOf(nVal))
			}
		}

		if fileInput, ok := v.Field(i).Interface().(FileUploadField); ok {
			files := form.File[field.Name]
			for _, file := range files {
				fd, err := file.Open()
				if err != nil {
					return err
				}

				fileInput.Files = append(fileInput.Files, ReceivedFile{
					Data: fd,
					Name: file.Filename,
					Size: file.Size,
				})
			}

			v.Field(i).Set(reflect.ValueOf(fileInput))
		}
	}

	return nil
}
