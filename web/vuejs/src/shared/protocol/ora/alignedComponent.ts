/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */
import type { Alignment } from '@/shared/protocol/ora/alignment';
import type { Component } from '@/shared/protocol/ora/component';

export interface AlignedComponent {
	// Component
	c /*omitempty*/? /*Component*/ : Component;

	/**
	 * Alignment may be empty and omitted. Then Center (=c) must be applied.
	 */
	// Alignment
	a /*omitempty*/? /*Alignment*/ : Alignment;
}
