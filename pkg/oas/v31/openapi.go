// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package oas

type OpenAPI struct {
	// Openapi is REQUIRED. This string MUST be the version number of the OpenAPI Specification that the OpenAPI Document uses. The openapi field SHOULD be used by tooling to interpret the OpenAPI Document. This is not related to the API info.version string.
	Openapi string `json:"openapi"`

	// Info is REQUIRED. Provides metadata about the API. The metadata MAY be used by tooling as required.
	Info Info `json:"info"`

	// The available paths and operations for the API. Required.
	Paths Paths `json:"paths"`

	// Tags is a list of tags used by the OpenAPI Description with additional metadata. The order of the tags can be used to reflect on their order by the parsing tools. Not all tags that are used by the Operation Object must be declared. The tags that are not declared MAY be organized randomly or based on the tools' logic. Each tag name in the list MUST be unique.
	Tags []Tag `json:"tags,omitempty"`

	// JsonSchemaDialect is the default value for the $schema keyword within Schema Objects contained within this OAS document. This MUST be in the form of a URI.
	JsonSchemaDialect string `json:"jsonSchemaDialect,omitempty"`

	// Additional external documentation.
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`

	// Components is an element to hold various Objects for the OpenAPI Description.
	Components *Components `json:"components,omitempty"`

	// Security is a declaration of which security mechanisms can be used across the API. The list of values includes alternative Security Requirement Objects that can be used. Only one of the Security Requirement Objects need to be satisfied to authorize a request. Individual operations can override this definition. The list can be incomplete, up to being empty or absent. To make security explicitly optional, an empty security requirement ({}) can be included in the array.
	Security *SecurityRequirement `json:"security,omitempty"`
}

func (api *OpenAPI) ComponentsSchemas() map[string]*Schema {
	if api.Components == nil {
		api.Components = &Components{}
	}

	if api.Components.Schemas == nil {
		api.Components.Schemas = map[string]*Schema{}
	}

	return api.Components.Schemas
}

func (api *OpenAPI) ComponentsSecurity() map[string]*SecurityScheme {
	if api.Components == nil {
		api.Components = &Components{}
	}

	if api.Components.SecuritySchemes == nil {
		api.Components.SecuritySchemes = map[string]*SecurityScheme{}
	}

	return api.Components.SecuritySchemes
}

func (api *OpenAPI) AddSchemaIfUnknown(s *Schema) {
	if s == nil {
		return
	}

	if s.Type == "" && s.Ref == "" {
		return
	}

	if s.GoName == "" {
		return
	}

	m := api.ComponentsSchemas()
	if _, ok := m[s.RefPlainName()]; ok {
		return
	}

	m[s.RefPlainName()] = s
}
