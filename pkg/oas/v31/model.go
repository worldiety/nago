// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package oas

const Version = "3.1.1"

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

type Operation struct {
	Tags         []string               `json:"tags,omitempty"`
	Summary      string                 `json:"summary,omitempty"`
	Description  string                 `json:"description,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`
	OperationID  string                 `json:"operationId,omitempty"`
	Parameters   []Parameter            `json:"parameters,omitempty"`
	RequestBody  *RequestBody           `json:"requestBody,omitempty"`
	Responses    Responses              `json:"responses,omitempty"`
	Deprecated   bool                   `json:"deprecated,omitempty"`

	// A declaration of which security mechanisms can be used across the API. The list of values includes alternative Security Requirement Objects that can be used. Only one of the Security Requirement Objects need to be satisfied to authorize a request. Individual operations can override this definition. The list can be incomplete, up to being empty or absent. To make security explicitly optional, an empty security requirement ({}) can be included in the array.
	Security []*SecurityRequirement `json:"security,omitempty"`
}

type SecurityRequirement map[string][]string

// Responses is a container for the expected responses of an operation. The container maps a HTTP response code to the expected response.
//
// The documentation is not necessarily expected to cover all possible HTTP response codes because they may not be known in advance. However, documentation is expected to cover a successful operation response and any known errors.
//
// The default MAY be used as a default Response Object for all HTTP codes that are not covered individually by the Responses Object.
//
// The Responses Object MUST contain at least one response code, and if only one response code is provided it SHOULD be the response for a successful operation call.
type Responses map[HttpStatusOrDefault]*Response

type Response struct {
	Description string               `json:"description,omitempty"`
	Headers     map[string]string    `json:"headers,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// HttpStatusOrDefault is for example, 2XX represents all response codes between 200 and 299. Only the following range definitions are allowed: 1XX, 2XX, 3XX, 4XX, and 5XX.
// Or just "default"
type HttpStatusOrDefault = string

type MediaTypeRange = string

// RequestBody describes a single request body.
type RequestBody struct {
	// A brief description of the request body. This could contain examples of use. [CommonMark] syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`

	// REQUIRED. The content of the request body. The key is a media type or media type range, see [RFC7231] Appendix D, and the value describes it. For requests that match multiple keys, only the most specific key is applicable. e.g. "text/plain" overrides "text/*"
	Content map[MediaTypeRange]MediaType `json:"content"`

	// Determines if the request body is required in the request. Defaults to false.
	Required bool `json:"required,omitempty"`
}

// MediaType provides schema and examples for the media type identified by its key.
//
// When example or examples are provided, the example SHOULD match the specified schema and be in the correct format as specified by the media type and its encoding. The example and examples fields are mutually exclusive, and if either is present it SHALL override any example in the schema. See Working With Examples for further guidance regarding the different ways of specifying examples, including non-JSON/YAML values.
type MediaType struct {
	// The schema defining the content of the request, response, parameter, or header.
	Schema *Schema `json:"schema,omitempty"`
	// Example of the media type; see Working With Examples.
	Example any `json:"example,omitempty"`
	// Examples of the media type; see Working With Examples.
	Examples map[MediaTypeRange]any `json:"examples,omitempty"`
	// A map between a property name and its encoding information. The key, being the property name, MUST exist in the schema as a property. The encoding field SHALL only apply to Request Body Objects, and only when the media type is multipart or application/x-www-form-urlencoded. If no Encoding Object is provided for a property, the behavior is determined by the default values documented for the Encoding Object.
	Encoding map[MediaTypeRange]Encoding `json:"encoding,omitempty"`
}

type Encoding struct{}

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
	//  	The schema defining the type used for the parameter.
	Schema   *Schema        `json:"schema,omitempty"`
	Examples map[string]any `json:"examples,omitempty"`
	Example  string         `json:"example,omitempty"`
}

type Location string

const (
	LocationHeader Location = "header"
	LocationCookie Location = "cookie"
	LocationPath   Location = "path"
	LocationQuery  Location = "query"
)
