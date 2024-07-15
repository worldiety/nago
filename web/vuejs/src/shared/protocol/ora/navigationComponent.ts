/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */


import type { Alignment } from '@/shared/protocol/ora/alignment';
import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { MenuEntry } from '@/shared/protocol/ora/menuEntry';
import type { Property } from '@/shared/protocol/ora/property';
import type { Ptr } from '@/shared/protocol/ora/ptr';
import type { SVG } from '@/shared/protocol/ora/sVG';

export interface NavigationComponent {
    // Ptr
    id /*Ptr*/: Ptr;
    // Type
    type: 'NavigationComponent'/*ComponentType*/;
    // Logo
    logo: Property<SVG>;
    // Menu
    menu: Property<MenuEntry[]>;
    // Alignment
    alignment: Property<Alignment>;
}

