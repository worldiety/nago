/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */
import type { Alignment } from '@/shared/protocol/ora/alignment';
import type { Component } from '@/shared/protocol/ora/component';
import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { Padding } from '@/shared/protocol/ora/padding';

/**
 * GridCell is undefined, if explicit row start/col start etc. is set and span values.
 */
export interface GridCell {
	// Type
	type: 'C' /*ComponentType*/;
	// Body
	b /*omitempty*/? /*Body*/ : Component;
	// ColStart
	cs /*omitempty*/? /*ColStart*/ : number /*int64*/;
	// ColEnd
	ce /*omitempty*/? /*ColEnd*/ : number /*int64*/;
	// RowStart
	rs /*omitempty*/? /*RowStart*/ : number /*int64*/;
	// RowEnd
	re /*omitempty*/? /*RowEnd*/ : number /*int64*/;
	// ColSpan
	cp /*omitempty*/? /*ColSpan*/ : number /*int64*/;
	// RowSpan
	rp /*omitempty*/? /*RowSpan*/ : number /*int64*/;
	// Padding
	p /*omitempty*/? /*Padding*/ : Padding;
	// Alignment
	a /*omitempty*/? /*Alignment*/ : Alignment;
}
