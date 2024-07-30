/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */


import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { Ptr } from '@/shared/protocol/ora/ptr';

/**
 * Radiobutton represents a user interface element which spans a visible area to click or tap from the user.
 * Usually a radiobutton belongs to a group, where only a single element can be picked. Thus, it is quite similar
 * to a Spinner/Select/Combobox.
 */
export interface Radiobutton {
    // Type
    type: 'R'/*ComponentType*/;

    /**
     * Value is the initial checked value.
     */
    // Value
    v/*omitempty*/? /*Value*/: boolean;

    /**
     * InputValue is where updated value of the checked states are written.
     */
    // InputValue
    i/*omitempty*/? /*InputValue*/: Ptr;
    // Disabled
    d/*omitempty*/? /*Disabled*/: boolean;
    // Invisible
    iv/*omitempty*/? /*Invisible*/: boolean;
}

