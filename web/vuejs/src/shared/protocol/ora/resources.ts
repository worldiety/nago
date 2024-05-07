// Code generated by github.com/worldiety/macro. DO NOT EDIT.


import type { RIDSVG } from '@/shared/protocol/ora/rIDSVG';
import type { SVG } from '@/shared/protocol/ora/sVG';

export class Resources {
    private _svgs : Record<RIDSVG,SVG>;
    get sVG(): Record<RIDSVG,SVG>{
        return this._svgs;
    }
    set sVG(value: Record<RIDSVG,SVG>){
        this._svgs = value;
    }
}

