/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */


import type { Component } from '@/shared/protocol/ora/component';
import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { Property } from '@/shared/protocol/ora/property';
import type { Ptr } from '@/shared/protocol/ora/ptr';

export interface TableCell {
    id /*Ptr*/: Ptr;
    type: 'TableCell'/*ComponentType*/;
    body: Property<Component>;
}

