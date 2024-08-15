/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */


import type { Border } from '@/shared/protocol/ora/border';
import type { Color } from '@/shared/protocol/ora/color';
import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { Font } from '@/shared/protocol/ora/font';
import type { Frame } from '@/shared/protocol/ora/frame';
import type { Padding } from '@/shared/protocol/ora/padding';
import type { Ptr } from '@/shared/protocol/ora/ptr';
import type { TextAlignment } from '@/shared/protocol/ora/textAlignment';

export interface Text {
    // Type
    type: 'T'/*ComponentType*/;
    // Value
    v/*omitempty*/? /*Value*/: string;

    /**
     * Color denotes the text color. Leave empty, for the context sensitiv default theme color.
     */
    // Color
    c/*omitempty*/? /*Color*/: Color;

    /**
     * BackgroundColor denotes the color of the text background.  Leave empty, for the context sensitiv default theme color.
     */
    // BackgroundColor
    bgc/*omitempty*/? /*BackgroundColor*/: Color;
    // OnClick
    onClick/*omitempty*/? /*OnClick*/: Ptr;
    // OnHoverStart
    onHoverStart/*omitempty*/? /*OnHoverStart*/: Ptr;
    // OnHoverEnd
    onHoverEnd/*omitempty*/? /*OnHoverEnd*/: Ptr;
    // Invisible
    i/*omitempty*/? /*Invisible*/: boolean;
    // Border
    b/*omitempty*/? /*Border*/: Border;
    // Padding
    p/*omitempty*/? /*Padding*/: Padding;
    // Frame
    f/*omitempty*/? /*Frame*/: Frame;

    /**
     * see also https://www.w3.org/WAI/tutorials/images/decision-tree/ but makes probably no sense.
     */
    // AccessibilityLabel
    al/*omitempty*/? /*AccessibilityLabel*/: string;
    // Font
    o/*omitempty*/? /*Font*/: Font;
    // Action
    t/*omitempty*/? /*Action*/: Ptr;
    // TextAlignment
    a/*omitempty*/? /*TextAlignment*/: TextAlignment;
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

