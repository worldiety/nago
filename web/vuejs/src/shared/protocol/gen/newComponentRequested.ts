// DO NOT EDIT. Generated by ora-gen-ts

import type { ComponentFactoryId } from '@/shared/protocol/componentFactoryId';
import type { RequestId } from '@/shared/protocol/requestId';


export interface NewComponentRequested {
    type: 'NewComponentRequested';
    
     /**
     * This locale has been picked by the backend.
     */
    activeLocale: string;
    
     /**
     * This is the unique address for a specific component factory, e.g. my/component/path. This is typically a page.
     */
    factory: ComponentFactoryId;
    
     /**
     * Contains string encoded parameters for a component. This is like query parameters.
     */
    values: Record<string, string>;
    
     /**
     * Request ID used to generate a new component request and is returned in the according response.
     */
    r: RequestId;
    
}
