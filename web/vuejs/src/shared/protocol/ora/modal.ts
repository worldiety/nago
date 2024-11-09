/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */


import type { Component } from '@/shared/protocol/ora/component';
import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { Ptr } from '@/shared/protocol/ora/ptr';

/**
 * A Modal can be declared at any place in the composed view tree. However, these dialogs are teleported into
 * the modal space in tree declaration order. A Modal is layouted above all other regular content and will
 * catch focus and disable controls of the views behind. Its bounds are at most the maximum possible screen size.
 */
export interface Modal {
    // Type
    type: 'M'/*ComponentType*/;
    // Content
    b/*omitempty*/? /*Content*/: Component;

    /**
     * OnDismissRequest is called, if the user wants to dismiss the dialog, e.g. by clicking outside or pressing escape.
     * You can then decide to disable you dialog, or not.
     */
    // OnDismissRequest
    odr/*omitempty*/? /*OnDismissRequest*/: Ptr;

	// ModalType 0==Dialog 1==overlay
	t/*omitempty*/? /*ModalType*/: number;

	//Top              Length    `json:"u"`
	u?:string;
	//Left             Length    `json:"l"`
	l?:string;
	//Right            Length    `json:"r"`
	r?:string;
	//Bottom           Length    `json:"bt,omitempty"`
	bt?:string;
}

