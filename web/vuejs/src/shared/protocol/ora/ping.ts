// Code generated by github.com/worldiety/macro. DO NOT EDIT.


import type { EventType } from '@/shared/protocol/ora/eventType';

export class Ping {
    private _type : EventType;
    get type(): EventType{
        return this._type;
    }
    set type(value: EventType){
        this._type = value;
    }
}

