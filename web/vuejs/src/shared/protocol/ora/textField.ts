/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */


import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { Property } from '@/shared/protocol/ora/property';
import type { Ptr } from '@/shared/protocol/ora/ptr';

export interface TextField {
    id /*Ptr*/: Ptr;
    type: 'TextField'/*ComponentType*/;
    label: Property<string>;
    hint: Property<string>;
    help: Property<string>;
    error: Property<string>;
    value: Property<string>;
    placeholder: Property<string>;
    disabled: Property<boolean>;
    simple: Property<boolean>;
    onTextChanged: Property<Ptr>;
    visible: Property<boolean>;
}

