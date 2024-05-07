// Code generated by github.com/worldiety/macro. DO NOT EDIT.


import type { Property } from '@/shared/protocol/ora/property';
import type { Component } from '@/shared/protocol/ora/component';
import type { Ptr } from '@/shared/protocol/ora/ptr';
import type { ComponentType } from '@/shared/protocol/ora/componentType';

export class VBox {
    private _id : Ptr;
    private _type : ComponentType;
    private _children : Property<Component[]>;
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
    get children(): Property<Component[]>{
        return this._children;
    }
    set children(value: Property<Component[]>){
        this._children = value;
    }
}

