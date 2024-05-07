// Code generated by github.com/worldiety/macro. DO NOT EDIT.


import type { Property } from '@/shared/protocol/ora/property';
import type { GridCell } from '@/shared/protocol/ora/gridCell';
import type { Ptr } from '@/shared/protocol/ora/ptr';
import type { ComponentType } from '@/shared/protocol/ora/componentType';

// TODO this entire type is so HTML like and hard to handle and port to mobile devices. It has no semantics.
// 
// I vote for deletion, but what is the replacement?
// 
// deprecated
export class Grid {
    private _id : Ptr;
    private _type : ComponentType;
    private _cells : Property<GridCell[]>;
    private _rows : Property<number>;
    private _columns : Property<number>;
    private _smColumns : Property<number>;
    private _mdColumns : Property<number>;
    private _lgColumns : Property<number>;
    private _gap : Property<string>;
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
    get cells(): Property<GridCell[]>{
        return this._cells;
    }
    set cells(value: Property<GridCell[]>){
        this._cells = value;
    }
    get rows(): Property<number>{
        return this._rows;
    }
    set rows(value: Property<number>){
        this._rows = value;
    }
    get columns(): Property<number>{
        return this._columns;
    }
    set columns(value: Property<number>){
        this._columns = value;
    }
    get sMColumns(): Property<number>{
        return this._smColumns;
    }
    set sMColumns(value: Property<number>){
        this._smColumns = value;
    }
    get mDColumns(): Property<number>{
        return this._mdColumns;
    }
    set mDColumns(value: Property<number>){
        this._mdColumns = value;
    }
    get lGColumns(): Property<number>{
        return this._lgColumns;
    }
    set lGColumns(value: Property<number>){
        this._lgColumns = value;
    }
    get gap(): Property<string>{
        return this._gap;
    }
    set gap(value: Property<string>){
        this._gap = value;
    }
}

