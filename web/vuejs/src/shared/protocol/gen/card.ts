// DO NOT EDIT. Generated by ora-gen-ts

import type { Pointer } from '@/shared/protocol/pointer';
import type { Property } from '@/shared/protocol/property';
import type { Component } from '@/shared/protocol/gen/component';


export interface Card {
    id: Pointer;
    type: 'Card';
    children: Property<Component[]>;
    action: Property<Pointer>;
    
}
