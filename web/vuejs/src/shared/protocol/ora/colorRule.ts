/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */


import type { Color } from '@/shared/protocol/ora/color';

export interface ColorRule {
    // Name
    name: string;

    /**
     * Light is the value for the user preferred color light mode
     */
    // Light
    light: Color;

    /**
     * Dark is the value for the user preferred color dark mode
     */
    // Dark
    dark: Color;
}

