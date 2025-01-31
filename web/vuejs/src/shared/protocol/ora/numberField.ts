/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */
import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { Property } from '@/shared/protocol/ora/property';
import type { Ptr } from '@/shared/protocol/ora/ptr';

export interface NumberField {
	// Ptr
	id /*Ptr*/ : Ptr;
	// Type
	type: 'NumberField' /*ComponentType*/;
	// Label
	label: Property<string>;
	// Hint
	hint: Property<string>;
	// Error
	error: Property<string>;
	// Value
	value: Property<string>;
	// Placeholder
	placeholder: Property<string>;
	// Disabled
	disabled: Property<boolean>;
	// Simple
	simple: Property<boolean>;
	// OnValueChanged
	onValueChanged: Property<Ptr>;
	// Visible
	visible: Property<boolean>;
}
