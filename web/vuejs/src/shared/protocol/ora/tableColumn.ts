/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */


import type { Alignment } from '@/shared/protocol/ora/alignment';
import type { Color } from '@/shared/protocol/ora/color';
import type { Component } from '@/shared/protocol/ora/component';
import type { Length } from '@/shared/protocol/ora/length';
import type { Ptr } from '@/shared/protocol/ora/ptr';

export interface TableColumn {
    // Content
    c/*omitempty*/? /*Content*/: Component;

    /**
     * Values higher than 1000 are clipped.
     */
    // ColSpan
    cs/*omitempty*/? /*ColSpan*/: number /*int*/;
    // Width
    w/*omitempty*/? /*Width*/: Length;
    // Alignment
    a/*omitempty*/? /*Alignment*/: Alignment;
    // BackgroundColor
    b/*omitempty*/? /*BackgroundColor*/: Color;
    // CellAction
    t/*omitempty*/? /*CellAction*/: Ptr;
}

