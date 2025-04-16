// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package oas

type Info struct {
	// Title is REQUIRED. The title of the API.
	Title string `json:"title"`

	// Summary is a short summary of the API.
	Summary string `json:"summary"`

	// Description is a description of the API. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description"`

	// TermsOfService is a URI for the Terms of Service for the API. This MUST be in the form of a URI.
	TermsOfService string `json:"termsOfService"`

	// Version is REQUIRED. The version of the OpenAPI Document (which is distinct from the OpenAPI Specification version or the version of the API being described or the version of the OpenAPI Description).
	Version string `json:"version"`

	// The contact information for the exposed API.
	Contact *Contact `json:"contact,omitempty"`

	// The license information for the exposed API.
	License *License `json:"license,omitempty"`
}

type Contact struct {
	// The identifying name of the contact person/organization.
	Name string `json:"name"`
	// The URL for the contact information. This MUST be in the form of a URL.
	URL string `json:"url"`
	// The email address of the contact person/organization. This MUST be in the form of an email address.
	Email string `json:"email"`
}

type License struct {
	// Name is REQUIRED. The license name used for the API.
	Name string `json:"name"`

	// URL is a URI for the license used for the API. This MUST be in the form of a URI. The url field is mutually exclusive of the identifier field.
	URL string `json:"url"`

	// Identifier is a SPDX license expression for the API. The identifier field is mutually exclusive of the url field.
	Identifier string `json:"identifier"`
}
