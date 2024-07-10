/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */


import type { Alignment } from '@/shared/protocol/ora/alignment';
import type { Component } from '@/shared/protocol/ora/component';
import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { Frame } from '@/shared/protocol/ora/frame';
import type { Length } from '@/shared/protocol/ora/length';
import type { NamedColor } from '@/shared/protocol/ora/namedColor';
import type { Padding } from '@/shared/protocol/ora/padding';

/**
 * An HStack aligns children elements in a horizontal row.
 * - the intrinsic component dimensions are the sum of all sizes of the contained children
 * - the parent can define a custom width and height
 * - if the container is larger than the contained views, it must center vertical or horizontal
 * - the inner gap between components should be around 2dp
 */
export interface HStack {
    type: 'hs'/*ComponentType*/;
    c/*omitempty*/? /*Children*/: Component[];

    /**
     * InnerGap is omitted, if empty
     */
    g/*omitempty*/? /*Gap*/: Length;

    /**
     * Frame is omitted if empty
     */
    f/*omitempty*/? /*Frame*/: Frame;

    /**
     * Alignment may be empty and omitted. Then Center (=c) must be applied.
     */
    a/*omitempty*/? /*Alignment*/: Alignment;

    /**
     * BackgroundColor regular is always transparent
     */
    bgc/*omitempty*/? /*BackgroundColor*/: NamedColor;
    p/*omitempty*/? /*Padding*/: Padding;
}

