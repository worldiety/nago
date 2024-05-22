/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */


import type { Component } from '@/shared/protocol/ora/component';
import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { ElementSize } from '@/shared/protocol/ora/elementSize';
import type { FlexAlignment } from '@/shared/protocol/ora/flexAlignment';
import type { Orientation } from '@/shared/protocol/ora/orientation';
import type { Property } from '@/shared/protocol/ora/property';
import type { Ptr } from '@/shared/protocol/ora/ptr';

export interface FlexContainer {
    id /*Ptr*/: Ptr;
    type: 'FlexContainer'/*ComponentType*/;
    elements: Property<Component[]>;
    elementSize: Property<ElementSize>;
    orientation: Property<Orientation>;
    contentAlignment: Property<FlexAlignment>;
    itemsAlignment: Property<FlexAlignment>;
}

