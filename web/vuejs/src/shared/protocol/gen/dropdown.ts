// DO NOT EDIT. Generated by ora-gen-ts

import type { Pointer } from '@/shared/protocol/pointer';
import type { Property } from '@/shared/protocol/property';
import type { DropdownItem } from '@/shared/protocol/gen/dropdownItem';


export interface Dropdown {
    id: Pointer;
    type: 'Dropdown';
    items: Property<DropdownItem[]>;
    selectedIndices: Property<number[]>;
    multiselect: Property<boolean>;
    expanded: Property<boolean>;
    disabled: Property<boolean>;
    label: Property<string>;
    hint: Property<string>;
    error: Property<string>;
    onClicked: Property<Pointer>;
    
}
