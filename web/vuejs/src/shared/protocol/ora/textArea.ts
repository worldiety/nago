// Code generated by github.com/worldiety/macro. DO NOT EDIT.


import type { Property } from '@/shared/protocol/ora/property';
import type { Ptr } from '@/shared/protocol/ora/ptr';
import type { ComponentType } from '@/shared/protocol/ora/componentType';

export class TextArea {
    private _id : Ptr;
    private _type : ComponentType;
    private _label : Property<string>;
    private _hint : Property<string>;
    private _error : Property<string>;
    private _value : Property<string>;
    private _rows : Property<number>;
    private _disabled : Property<boolean>;
    private _onTextChanged : Property<Ptr>;
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
    get value(): Property<string>{
        return this._value;
    }
    set value(value: Property<string>){
        this._value = value;
    }
    get rows(): Property<number>{
        return this._rows;
    }
    set rows(value: Property<number>){
        this._rows = value;
    }
    get disabled(): Property<boolean>{
        return this._disabled;
    }
    set disabled(value: Property<boolean>){
        this._disabled = value;
    }
    get onTextChanged(): Property<Ptr>{
        return this._onTextChanged;
    }
    set onTextChanged(value: Property<Ptr>){
        this._onTextChanged = value;
    }
}

