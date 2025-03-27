// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package oas

const Version = "3.0.4"

type OpenAPI struct {
	// This string MUST be the version number of the OpenAPI Specification that the OpenAPI document uses. The openapi field SHOULD be used by tooling to interpret the OpenAPI document. Required.
	// E.g. 3.1.0
	Openapi string `json:"openapi"`

	// Provides metadata about the API. The metadata MAY be used by tooling as required. Required.
	Info Info `json:"info"`

	// The available paths and operations for the API. Required.
	Paths Paths `json:"paths"`

	Tags []Tag `json:"tags"`

	// Additional external documentation.
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`
}

type Tag struct {
	// REQUIRED. The name of the tag.
	Name string `json:"name"`
	// A description for the tag. [CommonMark] syntax MAY be used for rich text representation.
	Description string `json:"description"`
	//  	Additional external documentation for this tag.
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`
}

type ExternalDocumentation struct {
	// A description of the target documentation. [CommonMark] syntax MAY be used for rich text representation.
	Description string `json:"description"`
	// REQUIRED. The URL for the target documentation. This MUST be in the form of a URL.
	URL string `json:"url"`
}

type Info struct {
	// The title of the API. Required.
	Title string `json:"title"`
	// The version of the OpenAPI document (which is distinct from the OpenAPI Specification version or the API implementation version). Required.
	Version string `json:"version"`
	// The contact information for the exposed API.
	Contact *Contact `json:"contact,omitempty"`
}

type Contact struct {
	// The identifying name of the contact person/organization.
	Name string `json:"name"`
	// The URL for the contact information. This MUST be in the form of a URL.
	URL string `json:"url"`
	// The email address of the contact person/organization. This MUST be in the form of an email address.
	Email string `json:"email"`
}

type Path = string

type Paths map[Path]*PathItem

type PathItem struct {
	// Allows for a referenced definition of this path item. The value MUST be in the form of a URL, and the referenced
	//structure MUST be in the form of a Path Item Object. In case a Path Item Object field appears both in
	//the defined object and the referenced object, the behavior is undefined. See the rules for resolving Relative References.
	Ref string `json:"$ref"`

	// An optional string summary, intended to apply to all operations in this path.
	Summary string `json:"summary"`

	// An optional string description, intended to apply to all operations in this path.
	// [CommonMark] syntax MAY be used for rich text representation.
	Description string `json:"description"`

	Get     *Operation `json:"get,omitempty"`
	Put     *Operation `json:"put,omitempty"`
	Post    *Operation `json:"post,omitempty"`
	Delete  *Operation `json:"delete,omitempty"`
	Patch   *Operation `json:"patch,omitempty"`
	Head    *Operation `json:"head,omitempty"`
	Options *Operation `json:"options,omitempty"`
	Trace   *Operation `json:"trace,omitempty"`

	Parameters []Parameter
}

type Operation struct {
	Tags         []string               `json:"tags"`
	Summary      string                 `json:"summary"`
	Description  string                 `json:"description"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`
	OperationID  string                 `json:"operationId"`
	Parameters   []Parameter            `json:"parameters"`
	RequestBody  *RequestBody           `json:"requestBody,omitempty"`
	Responses    Responses              `json:"responses"`
	Deprecated   bool                   `json:"deprecated"`
	Security     *SecurityRequirement   `json:"security,omitempty"`
}

type SecurityRequirement struct{}

// Responses is a container for the expected responses of an operation. The container maps a HTTP response code to the expected response.
//
// The documentation is not necessarily expected to cover all possible HTTP response codes because they may not be known in advance. However, documentation is expected to cover a successful operation response and any known errors.
//
// The default MAY be used as a default Response Object for all HTTP codes that are not covered individually by the Responses Object.
//
// The Responses Object MUST contain at least one response code, and if only one response code is provided it SHOULD be the response for a successful operation call.
type Responses map[HttpStatusOrDefault]*Response

type Response struct {
	Description string               `json:"description"`
	Headers     map[string]string    `json:"headers"`
	Content     map[string]MediaType `json:"content"`
}

// HttpStatusOrDefault is for example, 2XX represents all response codes between 200 and 299. Only the following range definitions are allowed: 1XX, 2XX, 3XX, 4XX, and 5XX.
// Or just "default"
type HttpStatusOrDefault = string

type MediaTypeRange = string

// RequestBody describes a single request body.
type RequestBody struct {
	// A brief description of the request body. This could contain examples of use. [CommonMark] syntax MAY be used for rich text representation.
	Description string `json:"description"`

	// REQUIRED. The content of the request body. The key is a media type or media type range, see [RFC7231] Appendix D, and the value describes it. For requests that match multiple keys, only the most specific key is applicable. e.g. "text/plain" overrides "text/*"
	Content map[MediaTypeRange]MediaType

	// Determines if the request body is required in the request. Defaults to false.
	Required bool `json:"required"`
}

// MediaType provides schema and examples for the media type identified by its key.
//
// When example or examples are provided, the example SHOULD match the specified schema and be in the correct format as specified by the media type and its encoding. The example and examples fields are mutually exclusive, and if either is present it SHALL override any example in the schema. See Working With Examples for further guidance regarding the different ways of specifying examples, including non-JSON/YAML values.
type MediaType struct {
	// The schema defining the content of the request, response, parameter, or header.
	Schema Schema `json:"schema"`
	// Example of the media type; see Working With Examples.
	Example any `json:"example"`
	// Examples of the media type; see Working With Examples.
	Examples map[MediaTypeRange]any `json:"examples"`
	// A map between a property name and its encoding information. The key, being the property name, MUST exist in the schema as a property. The encoding field SHALL only apply to Request Body Objects, and only when the media type is multipart or application/x-www-form-urlencoded. If no Encoding Object is provided for a property, the behavior is determined by the default values documented for the Encoding Object.
	Encoding map[MediaTypeRange]Encoding `json:"encoding"`
}

type Encoding struct{}

type Schema struct {
	Ref string `json:"$ref"`
}

type Parameter struct {
	// REQUIRED. The name of the parameter. Parameter names are case sensitive.
	//
	//    If in is "path", the name field MUST correspond to a template expression occurring within the path field in the Paths Object. See Path Templating for further information.
	//    If in is "header" and the name field is "Accept", "Content-Type" or "Authorization", the parameter definition SHALL be ignored.
	//    For all other cases, the name corresponds to the parameter name used by the in field.
	Name string `json:"name"`
	// REQUIRED. The location of the parameter. Possible values are "query", "header", "path" or "cookie".
	In Location `json:"in"`

	// A brief description of the parameter. This could contain examples of use. [CommonMark] syntax MAY be used for rich text representation.
	Description string `json:"description"`

	// Determines whether this parameter is mandatory. If the parameter location is "path", this field is REQUIRED and its value MUST be true. Otherwise, the field MAY be included and its default value is false.
	Required bool `json:"required"`

	// Specifies that a parameter is deprecated and SHOULD be transitioned out of usage. Default value is false.
	Deprecated bool `json:"deprecated"`

	// If true, clients MAY pass a zero-length string value in place of parameters that would otherwise be omitted entirely, which the server SHOULD interpret as the parameter being unused. Default value is false. If style is used, and if behavior is n/a (cannot be serialized), the value of allowEmptyValue SHALL be ignored. Interactions between this field and the parameter’s Schema Object are implementation-defined. This field is valid only for query parameters. Use of this field is NOT RECOMMENDED, and it is likely to be removed in a later revision.
	AllowEmptyValue bool `json:"allowEmptyValue"`
}

type Location string

const (
	LocationHeader Location = "header"
	LocationCookie Location = "cookie"
	LocationPath   Location = "path"
	LocationQuery  Location = "query"
)
