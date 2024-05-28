/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */


import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { Property } from '@/shared/protocol/ora/property';
import type { Ptr } from '@/shared/protocol/ora/ptr';

export interface Text {
    id /*Ptr*/: Ptr;
    type: 'Text'/*ComponentType*/;
    value: Property<string>;
    color: Property<string>;
    colorDark: Property<string>;
    size: Property<string>;
    onClick: Property<Ptr>;
    onHoverStart: Property<Ptr>;
    onHoverEnd: Property<Ptr>;
    visible: Property<boolean>;
}

