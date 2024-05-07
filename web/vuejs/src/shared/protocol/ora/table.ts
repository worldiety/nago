// Code generated by github.com/worldiety/macro. DO NOT EDIT.


import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { Property } from '@/shared/protocol/ora/property';
import type { Ptr } from '@/shared/protocol/ora/ptr';
import type { TableCell } from '@/shared/protocol/ora/tableCell';
import type { TableRow } from '@/shared/protocol/ora/tableRow';

export class Table {
    private _id : Ptr;
    private _type : ComponentType;
    private _headers : Property<TableCell[]>;
    private _rows : Property<TableRow[]>;
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
    get headers(): Property<TableCell[]>{
        return this._headers;
    }
    set headers(value: Property<TableCell[]>){
        this._headers = value;
    }
    get rows(): Property<TableRow[]>{
        return this._rows;
    }
    set rows(value: Property<TableRow[]>){
        this._rows = value;
    }
}

