/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */


import type { DP } from '@/shared/protocol/ora/dP';
import type { Density } from '@/shared/protocol/ora/density';

/**
 * WindowInfo describes the area into which the frontend renders the ora view tree.
 * A user can simply change the layout of the screen, e.g. by rotation the smartphone or
 * changing the size of a browser window.
 */
export interface WindowInfo {
    width: DP;
    height: DP;
    density: Density;
}

