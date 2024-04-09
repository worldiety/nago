// DO NOT EDIT. Generated by ora-gen-ts

import type { RequestId } from '@/shared/protocol/requestId';


export interface Acknowledged {
    
     /**
     * The magic type constant.
     */
    type: 'A';
    
     /**
     * The request identifier, which is sent back.
     */
    r: RequestId;
    
}
