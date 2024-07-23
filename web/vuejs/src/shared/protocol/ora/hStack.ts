/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */


import type { Alignment } from '@/shared/protocol/ora/alignment';
import type { Border } from '@/shared/protocol/ora/border';
import type { Color } from '@/shared/protocol/ora/color';
import type { Component } from '@/shared/protocol/ora/component';
import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { Font } from '@/shared/protocol/ora/font';
import type { Frame } from '@/shared/protocol/ora/frame';
import type { Length } from '@/shared/protocol/ora/length';
import type { Padding } from '@/shared/protocol/ora/padding';
import type { Ptr } from '@/shared/protocol/ora/ptr';

/**
 * An HStack aligns children elements in a horizontal row.
 * - the intrinsic component dimensions are the sum of all sizes of the contained children
 * - the parent can define a custom width and height
 * - if the container is larger than the contained views, it must center vertical or horizontal
 * - the inner gap between components should be around 2dp
 */
export interface HStack {
    // Type
    type: 'hs'/*ComponentType*/;
    // Children
    c/*omitempty*/? /*Children*/: Component[];

    /**
     * InnerGap is omitted, if empty
     */
    // Gap
    g/*omitempty*/? /*Gap*/: Length;

    /**
     * Frame is omitted if empty
     */
    // Frame
    f/*omitempty*/? /*Frame*/: Frame;

    /**
     * Alignment may be empty and omitted. Then Center (=c) must be applied.
     */
    // Alignment
    a/*omitempty*/? /*Alignment*/: Alignment;
    // BackgroundColor
    bgc/*omitempty*/? /*BackgroundColor*/: Color;
    // Padding
    p/*omitempty*/? /*Padding*/: Padding;
    // Border
    b/*omitempty*/? /*Border*/: Border;

    /**
     * see also https://www.w3.org/WAI/tutorials/images/decision-tree/
     */
    // AccessibilityLabel
    al/*omitempty*/? /*AccessibilityLabel*/: string;
    // Invisible
    iv/*omitempty*/? /*Invisible*/: boolean;
    // Font
    fn/*omitempty*/? /*Font*/: Font;
    // Action
    t/*omitempty*/? /*Action*/: Ptr;
    // HoveredBackgroundColor
    hgc/*omitempty*/? /*HoveredBackgroundColor*/: Color;
    // PressedBackgroundColor
    pgc/*omitempty*/? /*PressedBackgroundColor*/: Color;
    // FocusedBackgroundColor
    fbc/*omitempty*/? /*FocusedBackgroundColor*/: Color;
    // HoveredBorder
    hb/*omitempty*/? /*HoveredBorder*/: Border;
    // PressedBorder
    pb/*omitempty*/? /*PressedBorder*/: Border;
    // FocusedBorder
    fb/*omitempty*/? /*FocusedBorder*/: Border;
}

