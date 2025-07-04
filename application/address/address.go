// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package address

type ID string

// The Address represents a physical or mailing address, including geolocation and metadata.
// It supports international address formats and can be used for home, work, billing, etc.
type Address struct {
	ID          ID
	Type        Type    `json:"type,omitempty"`        // The type of address, e.g., "home", "work", "billing", "shipping".
	Street      string  `json:"street,omitempty"`      // Street name and number, e.g., "Main Street 5".
	Street2     string  `json:"street2,omitempty"`     // Additional street info, such as apartment number or c/o.
	PostalCode  string  `json:"postalCode,omitempty"`  // ZIP or postal code.
	City        string  `json:"city,omitempty"`        // City or locality.
	State       string  `json:"state,omitempty"`       // State, province, or region (e.g., "California", "Bayern").
	Country     string  `json:"country,omitempty"`     // Full country name, e.g., "Germany".
	CountryCode string  `json:"countryCode,omitempty"` // ISO 3166-1 alpha-2 country code, e.g., "DE", "US".
	TimeZone    string  `json:"timeZone,omitempty"`    // IANA timezone name, e.g., "Europe/Berlin".
	Latitude    float64 `json:"latitude,omitempty"`    // Geographic latitude for geolocation or map features.
	Longitude   float64 `json:"longitude,omitempty"`   // Geographic longitude.
	IsPrimary   bool    `json:"isPrimary,omitempty"`   // Indicates if this is the primary or preferred address.
}

func (a Address) Identity() ID {
	return a.ID
}

func (a Address) WithIdentity(id ID) Address {
	a.ID = id
	return a
}

type Type string

const (
	Shipping Type = "shipping"
	Billing  Type = "billing"
	Work     Type = "work"
	Home     Type = "home"
)

func (a Address) IsZero() bool {
	return a == Address{}
}
