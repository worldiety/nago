// DO NOT EDIT. Generated by ora-gen-ts

import type { Pointer } from '@/shared/protocol/pointer';
import type { Property } from '@/shared/protocol/property';


export interface Slider {
    id: Pointer;
    type: 'Slider';
    disabled: Property<boolean>;
    label: Property<string>;
    hint: Property<string>;
    error: Property<string>;
    startValue: Property<number>;
    endValue: Property<number>;
    min: Property<number>;
    max: Property<number>;
    stepsize: Property<number>;
    startInitialized: Property<boolean>;
    endInitialized: Property<boolean>;
    showLabel: Property<boolean>;
    labelSuffix: Property<string>;
    onChanged: Property<Pointer>;
    
}
