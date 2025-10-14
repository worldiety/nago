// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package function

import "encoding/json"

type Function struct {
	Name        string      `json:"name,omitempty"`
	Description string      `json:"desc,omitempty"`
	Parameters  []Parameter `json:"params,omitempty"`
}

func (f Function) intoJSON() jsonSchemaType {
	spec := jsonSchemaType{
		Name:        f.Name,
		Type:        Func,
		Description: f.Description,
	}

	if len(f.Parameters) > 0 {
		tmpObj := Parameter{
			Type:       Object,
			Properties: f.Parameters,
		}

		params := tmpObj.intoJSON()
		spec.Parameters = &params
	}

	return spec
}

func (f Function) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.intoJSON())
}

type Parameter struct {
	Name        string      `json:"name,omitempty"`
	Type        Type        `json:"type,omitempty"`
	Description string      `json:"desc,omitempty"`
	Required    bool        `json:"required,omitempty"` // if not required, it gets the zero value.
	Properties  []Parameter `json:"properties,omitempty"`
}

func (p Parameter) intoJSON() jsonSchemaType {
	if len(p.Properties) > 0 {
		p.Type = Object
	}

	spec := jsonSchemaType{
		// omit the name
		Type:        p.Type,
		Description: p.Description,
	}

	if len(p.Properties) > 0 {
		spec.Properties = map[string]jsonSchemaType{}

		for _, property := range p.Properties {
			spec.Properties[property.Name] = property.intoJSON()
			if property.Required {
				spec.Required = append(spec.Required, property.Name)
			}
		}
	}

	return spec
}

// Type matches the json schema basic types.
type Type string

const (
	String  Type = "string"
	Integer Type = "integer"
	Float   Type = "number"
	Boolean Type = "boolean"
	Object  Type = "object"
	Null    Type = "null"
	Func    Type = "function"
)

// jsonSchemaType is a common format to describe function callings for OpenAI and MistralAI in 2025.
type jsonSchemaType struct {
	Type        Type                      `json:"type,omitempty"`
	Name        string                    `json:"name,omitempty"`
	Description string                    `json:"description,omitempty"`
	Parameters  *jsonSchemaType           `json:"parameters,omitempty"`
	Properties  map[string]jsonSchemaType `json:"properties,omitempty"`
	Required    []string                  `json:"required,omitempty"`
}
