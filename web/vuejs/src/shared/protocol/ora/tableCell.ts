// Code generated by github.com/worldiety/macro. DO NOT EDIT.


import type { Component } from '@/shared/protocol/ora/component';
import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { Property } from '@/shared/protocol/ora/property';
import type { Ptr } from '@/shared/protocol/ora/ptr';

export class TableCell {
    private _id : Ptr;
    private _type : ComponentType;
    private _body : Property<Component>;
    get ptr(): Ptr{
        return this._id;
    }
    set ptr(value: Ptr){
        this._id = value;
    }
    get type(): ComponentType{
        return this._type;
    }
    set type(value: ComponentType){
        this._type = value;
    }
    get body(): Property<Component>{
        return this._body;
    }
    set body(value: Property<Component>){
        this._body = value;
    }
}

