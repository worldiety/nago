// DO NOT EDIT. Generated by ora-gen-ts

import type { ComponentFactoryId } from '@/shared/protocol/componentFactoryId';
import type { RequestId } from '@/shared/protocol/requestId';



/**
* NewComponentRequested allocates an addressable component explicitely in the backend within its channel scope.
Adressable components are like pages in a classic server side rendering or like routing targets in single page apps.
We do not call them _page_ anymore, because that has wrong assocations in the web world.
Adressable components exist independently from each other and share no lifecycle with each other.
However, a frontend can create as many component instances it wants.
It does not matter, if these components are of the same type, addresses or entirely different.
The backend responds with a component invalidation event.

Factories of addressable components are always stateless.
However, often it does not make sense without additional parameters, e.g. because a detail view needs to know which entity has to be displayed.

*/
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
    values: Map<string, string>;
    
     /**
     * Request ID used to generate a new component request and is returned in the according response.
     */
    requestId: RequestId;
    
}
