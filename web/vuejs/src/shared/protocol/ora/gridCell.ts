// Code generated by github.com/worldiety/macro. DO NOT EDIT.


import type { Component } from '@/shared/protocol/ora/component';
import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { Property } from '@/shared/protocol/ora/property';
import type { Ptr } from '@/shared/protocol/ora/ptr';

// TODO this entire type is so HTML like and hard to handle and port to mobile devices. It has no semantics.
// 
// I vote for deletion, but what is the replacement?
// 
// deprecated
export class GridCell {
    private _id : Ptr;
    private _type : ComponentType;
    private _body : Property<Component>;
    private _colStart : Property<number>;
    private _colEnd : Property<number>;
    private _rowStart : Property<number>;
    private _rowEnd : Property<number>;
    private _colSpan : Property<number>;
    private _smColSpan : Property<number>;
    private _mdColSpan : Property<number>;
    private _lgColSpan : Property<number>;
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
    get colStart(): Property<number>{
        return this._colStart;
    }
    set colStart(value: Property<number>){
        this._colStart = value;
    }
    get colEnd(): Property<number>{
        return this._colEnd;
    }
    set colEnd(value: Property<number>){
        this._colEnd = value;
    }
    get rowStart(): Property<number>{
        return this._rowStart;
    }
    set rowStart(value: Property<number>){
        this._rowStart = value;
    }
    get rowEnd(): Property<number>{
        return this._rowEnd;
    }
    set rowEnd(value: Property<number>){
        this._rowEnd = value;
    }
    get colSpan(): Property<number>{
        return this._colSpan;
    }
    set colSpan(value: Property<number>){
        this._colSpan = value;
    }
    get smColSpan(): Property<number>{
        return this._smColSpan;
    }
    set smColSpan(value: Property<number>){
        this._smColSpan = value;
    }
    get mdColSpan(): Property<number>{
        return this._mdColSpan;
    }
    set mdColSpan(value: Property<number>){
        this._mdColSpan = value;
    }
    get lgColSpan(): Property<number>{
        return this._lgColSpan;
    }
    set lgColSpan(value: Property<number>){
        this._lgColSpan = value;
    }
}

