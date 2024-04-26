// DO NOT EDIT. Generated by ora-gen-ts

import type { Pointer } from '@/shared/protocol/pointer';
import type { RequestId } from '@/shared/protocol/requestId';


export interface SetPropertyValueRequested {
    
     /**
     * P stands for Set**P**ropertValue. It is expected, that we must process countless of these events.
     */
    type: 'P';
    
     /**
     * p denotes the remote pointer.
     */
    p: Pointer;
    
     /**
     * v denotes the serialized value to set the property to.
     */
    v: string;
    requestId?: RequestId;
    
}
