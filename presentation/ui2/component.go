package ui2

import (
	"fmt"
	"github.com/swaggest/openapi-go/openapi3"
	"regexp"
)

var validComponentIdRegex = regexp.MustCompile(`[a-z_\-0-9]+`)

type ComponentID string

func (p ComponentID) Validate() error {
	if p == "" {
		return fmt.Errorf("must not be empty")
	}

	if len(validComponentIdRegex.FindString(string(p))) != len(p) {
		return fmt.Errorf("the id '%s' is invalid and must match the [a-z_\\-0-9]+", string(p))
	}

	return nil
}

type Component[Params any] interface {
	ComponentID() ComponentID
	configure(parentSlug string, r router)
	renderOpenAPI(p Params, tag string, parentSlug string, r *openapi3.Reflector)
}

// TypeDiscriminator identifies a distinct type.
// Most distinct types within this API contain a type field to distinguish between various types.
// Otherwise, nested graphs of dynamic components can not be resolved properly.
type TypeDiscriminator string

// Link represents a follow-up action similar to the HATEOAS style.
type Link string

func (Link) Description() string {
	return "Link represents a follow-up action similar to the HATEOAS style."
}
