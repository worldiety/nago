// Code generated by github.com/worldiety/macro. DO NOT EDIT.


import type { EventType } from '@/shared/protocol/ora/eventType';
import type { RequestId } from '@/shared/protocol/ora/requestId';
import type { Component } from '@/shared/protocol/ora/component';

export class ComponentInvalidated {
    private _type : EventType;
    private _r : RequestId;
    private _value : Component;
    get type(): EventType{
        return this._type;
    }
    set type(value: EventType){
        this._type = value;
    }
    get requestId(): RequestId{
        return this._r;
    }
    set requestId(value: RequestId){
        this._r = value;
    }
    get component(): Component{
        return this._value;
    }
    set component(value: Component){
        this._value = value;
    }
}

