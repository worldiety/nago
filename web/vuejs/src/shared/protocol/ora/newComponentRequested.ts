// Code generated by github.com/worldiety/macro. DO NOT EDIT.


import type { ComponentFactoryId } from '@/shared/protocol/ora/componentFactoryId';
import type { RequestId } from '@/shared/protocol/ora/requestId';
import type { EventType } from '@/shared/protocol/ora/eventType';

// NewComponentRequested allocates an addressable component explicitely in the backend within its channel scope.
// Adressable components are like pages in a classic server side rendering or like routing targets in single page apps.
// We do not call them _page_ anymore, because that has wrong assocations in the web world.
// Adressable components exist independently from each other and share no lifecycle with each other.
// However, a frontend can create as many component instances it wants.
// It does not matter, if these components are of the same type, addresses or entirely different.
// The backend responds with a component invalidation event.
// 
// Factories of addressable components are always stateless.
// However, often it does not make sense without additional parameters, e.g. because a detail view needs to know which entity has to be displayed.
export class NewComponentRequested {
    private _type : EventType;
    private _activeLocale : string;
    private _factory : ComponentFactoryId;
    private _values : Record<string,string>;
    private _r : RequestId;
    get type(): EventType{
        return this._type;
    }
    set type(value: EventType){
        this._type = value;
    }
    get locale(): string{
        return this._activeLocale;
    }
    set locale(value: string){
        this._activeLocale = value;
    }
    get factory(): ComponentFactoryId{
        return this._factory;
    }
    set factory(value: ComponentFactoryId){
        this._factory = value;
    }
    get values(): Record<string,string>{
        return this._values;
    }
    set values(value: Record<string,string>){
        this._values = value;
    }
    get requestId(): RequestId{
        return this._r;
    }
    set requestId(value: RequestId){
        this._r = value;
    }
}

