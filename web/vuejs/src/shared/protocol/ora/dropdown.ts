// Code generated by github.com/worldiety/macro. DO NOT EDIT.


import type { Ptr } from '@/shared/protocol/ora/ptr';
import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { Property } from '@/shared/protocol/ora/property';
import type { DropdownItem } from '@/shared/protocol/ora/dropdownItem';

export class Dropdown {
    private _id : Ptr;
    private _type : ComponentType;
    private _items : Property<DropdownItem[]>;
    private _selectedIndices : Property<number[]>;
    private _multiselect : Property<boolean>;
    private _expanded : Property<boolean>;
    private _disabled : Property<boolean>;
    private _label : Property<string>;
    private _hint : Property<string>;
    private _error : Property<string>;
    private _onClicked : Property<Ptr>;
    private _searchable : Property<boolean>;
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
    get items(): Property<DropdownItem[]>{
        return this._items;
    }
    set items(value: Property<DropdownItem[]>){
        this._items = value;
    }
    get selectedIndices(): Property<number[]>{
        return this._selectedIndices;
    }
    set selectedIndices(value: Property<number[]>){
        this._selectedIndices = value;
    }
    get multiselect(): Property<boolean>{
        return this._multiselect;
    }
    set multiselect(value: Property<boolean>){
        this._multiselect = value;
    }
    get expanded(): Property<boolean>{
        return this._expanded;
    }
    set expanded(value: Property<boolean>){
        this._expanded = value;
    }
    get disabled(): Property<boolean>{
        return this._disabled;
    }
    set disabled(value: Property<boolean>){
        this._disabled = value;
    }
    get label(): Property<string>{
        return this._label;
    }
    set label(value: Property<string>){
        this._label = value;
    }
    get hint(): Property<string>{
        return this._hint;
    }
    set hint(value: Property<string>){
        this._hint = value;
    }
    get error(): Property<string>{
        return this._error;
    }
    set error(value: Property<string>){
        this._error = value;
    }
    get onClicked(): Property<Ptr>{
        return this._onClicked;
    }
    set onClicked(value: Property<Ptr>){
        this._onClicked = value;
    }
    get searchable(): Property<boolean>{
        return this._searchable;
    }
    set searchable(value: Property<boolean>){
        this._searchable = value;
    }
}

