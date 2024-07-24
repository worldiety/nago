/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */


import type { Border } from '@/shared/protocol/ora/border';
import type { Color } from '@/shared/protocol/ora/color';
import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { Frame } from '@/shared/protocol/ora/frame';
import type { Padding } from '@/shared/protocol/ora/padding';
import type { Ptr } from '@/shared/protocol/ora/ptr';
import type { URI } from '@/shared/protocol/ora/uRI';

export interface Image {
    // Type
    type: 'I'/*ComponentType*/;
    // URI
    u/*omitempty*/? /*URI*/: URI;

    /**
     * see also https://www.w3.org/WAI/tutorials/images/decision-tree/
     */
    // AccessibilityLabel
    al/*omitempty*/? /*AccessibilityLabel*/: string;
    // Invisible
    iv/*omitempty*/? /*Invisible*/: boolean;
    // Border
    b/*omitempty*/? /*Border*/: Border;
    // Frame
    f/*omitempty*/? /*Frame*/: Frame;
    // Padding
    p/*omitempty*/? /*Padding*/: Padding;
    // SVG
    s/*omitempty*/? /*SVG*/: 'no type resolved';
    // CachedSVG
    v/*omitempty*/? /*CachedSVG*/: Ptr;
    // FillColor
    c/*omitempty*/? /*FillColor*/: Color;
    // StrokeColor
    k/*omitempty*/? /*StrokeColor*/: Color;
}

