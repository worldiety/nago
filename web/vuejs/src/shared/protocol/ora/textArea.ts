/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */
import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { Property } from '@/shared/protocol/ora/property';
import type { Ptr } from '@/shared/protocol/ora/ptr';

export interface TextArea {
	// Ptr
	id /*Ptr*/ : Ptr;
	// Type
	type: 'TextArea' /*ComponentType*/;
	// Label
	label: Property<string>;
	// Hint
	hint: Property<string>;
	// Error
	error: Property<string>;
	// Value
	value: Property<string>;
	// Rows
	rows: Property<number /*int64*/>;
	// Disabled
	disabled: Property<boolean>;
	// OnTextChanged
	onTextChanged: Property<Ptr>;
	// Visible
	visible: Property<boolean>;
}
