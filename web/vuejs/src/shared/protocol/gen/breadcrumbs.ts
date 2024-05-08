// DO NOT EDIT. Generated by ora-gen-ts

import type { Pointer } from '@/shared/protocol/pointer';
import type { Property } from '@/shared/protocol/property';
import type { BreadcrumbItem } from '@/shared/protocol/gen/breadcrumbItem';
import type { SVG } from '@/shared/protocol/svg';


export interface Breadcrumbs {
    id: Pointer;
    type: 'Breadcrumbs';
    items: Property<BreadcrumbItem[]>;
    selectedItemIndex: Property<number>;
    icon: Property<SVG>;
    
}
