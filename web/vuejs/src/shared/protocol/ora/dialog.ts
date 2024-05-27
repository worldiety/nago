/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */


import type { Component } from '@/shared/protocol/ora/component';
import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { ElementSize } from '@/shared/protocol/ora/elementSize';
import type { Property } from '@/shared/protocol/ora/property';
import type { Ptr } from '@/shared/protocol/ora/ptr';
import type { SVG } from '@/shared/protocol/ora/sVG';

export interface Dialog {
    id /*Ptr*/: Ptr;
    type: 'Dialog'/*ComponentType*/;
    title: Property<string>;
    body: Property<Component>;
    footer: Property<Component>;
    icon: Property<SVG>;
    visible: Property<boolean>;
    timestamp: Property<number /*int64*/>;
    size: Property<ElementSize>;
}

