// Code generated by github.com/worldiety/macro. DO NOT EDIT.


import type { ComponentType } from '@/shared/protocol/ora/componentType';
import type { Property } from '@/shared/protocol/ora/property';
import type { Ptr } from '@/shared/protocol/ora/ptr';

export class Text {
    private _id : Ptr;
    private _type : ComponentType;
    private _value : Property<string>;
    private _color : Property<string>;
    private _colorDark : Property<string>;
    private _size : Property<string>;
    private _onClick : Property<Ptr>;
    private _onHoverStart : Property<Ptr>;
    private _onHoverEnd : Property<Ptr>;
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
    get value(): Property<string>{
        return this._value;
    }
    set value(value: Property<string>){
        this._value = value;
    }
    get color(): Property<string>{
        return this._color;
    }
    set color(value: Property<string>){
        this._color = value;
    }
    get colorDark(): Property<string>{
        return this._colorDark;
    }
    set colorDark(value: Property<string>){
        this._colorDark = value;
    }
    get size(): Property<string>{
        return this._size;
    }
    set size(value: Property<string>){
        this._size = value;
    }
    get onClick(): Property<Ptr>{
        return this._onClick;
    }
    set onClick(value: Property<Ptr>){
        this._onClick = value;
    }
    get onHoverStart(): Property<Ptr>{
        return this._onHoverStart;
    }
    set onHoverStart(value: Property<Ptr>){
        this._onHoverStart = value;
    }
    get onHoverEnd(): Property<Ptr>{
        return this._onHoverEnd;
    }
    set onHoverEnd(value: Property<Ptr>){
        this._onHoverEnd = value;
    }
}

