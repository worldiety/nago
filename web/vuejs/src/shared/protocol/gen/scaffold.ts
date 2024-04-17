// DO NOT EDIT. Generated by ora-gen-ts

import type { Pointer } from '@/shared/protocol/pointer';
import type { Property } from '@/shared/protocol/property';
import type { Component } from '@/shared/protocol/gen/component';
import type { Button } from '@/shared/protocol/gen/button';


export interface Scaffold {
    id: Pointer;
    type: 'Scaffold';
    title: Property<string>;
    body: Property<Component>;
    breadcrumbs: Property<Button[]>;
    menu: Property<Button[]>;
    topbarLeft: Property<Component>;
    topbarMid: Property<Component>;
    topbarRight: Property<Component>;
    
}
