/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */


import type { Color } from '@/shared/protocol/ora/color';
import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { Frame } from '@/shared/protocol/ora/frame';
import type { Length } from '@/shared/protocol/ora/length';
import type { Padding } from '@/shared/protocol/ora/padding';
import type { Ptr } from '@/shared/protocol/ora/ptr';

export interface Text {
    // Ptr
    id /*Ptr*/: Ptr;
    // Type
    type: 'Text'/*ComponentType*/;
    // Value
    value/*omitempty*/? /*Value*/: string;

    /**
     * Color denotes the text color. Leave empty, for the context sensitiv default theme color.
     */
    // Color
    color/*omitempty*/? /*Color*/: Color;

    /**
     * BackgroundColor denotes the color of the text background.  Leave empty, for the context sensitiv default theme color.
     */
    // BackgroundColor
    backgroundColor/*omitempty*/? /*BackgroundColor*/: Color;
    // Size
    s/*omitempty*/? /*Size*/: Length;
    // OnClick
    onClick/*omitempty*/? /*OnClick*/: Ptr;
    // OnHoverStart
    onHoverStart/*omitempty*/? /*OnHoverStart*/: Ptr;
    // OnHoverEnd
    onHoverEnd/*omitempty*/? /*OnHoverEnd*/: Ptr;
    // Invisible
    invisible/*omitempty*/? /*Invisible*/: boolean;
    // Padding
    p/*omitempty*/? /*Padding*/: Padding;
    // Frame
    f/*omitempty*/? /*Frame*/: Frame;

    /**
     * see also https://www.w3.org/WAI/tutorials/images/decision-tree/ but makes probably no sense.
     */
    // AccessibilityLabel
    al/*omitempty*/? /*AccessibilityLabel*/: string;
}

