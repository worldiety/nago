/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */


import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { Property } from '@/shared/protocol/ora/property';
import type { Ptr } from '@/shared/protocol/ora/ptr';

/**
 * Checkbox represents an user interface element which spans a visible area to click or tap from the user.
 */
export interface Checkbox {
    id /*Ptr*/: Ptr;
    type: 'Checkbox'/*ComponentType*/;
    selected: Property<boolean>;
    onClicked: Property<Ptr>;
    disabled: Property<boolean>;
}

