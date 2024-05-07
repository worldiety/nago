// Code generated by github.com/worldiety/macro. DO NOT EDIT.


import type { EventType } from '@/shared/protocol/ora/eventType';
import type { Themes } from '@/shared/protocol/ora/themes';
import type { Resources } from '@/shared/protocol/ora/resources';
import type { RequestId } from '@/shared/protocol/ora/requestId';

// A ConfigurationDefined event is the response to a [ConfigurationRequested] event.
// According to the locale request, string and svg resources can be localized by the backend.
// The returned locale is the actually picked locale from the requested locale query string.
// 
// It looks quite obfuscated, however this minified version is intentional, because it may succeed each transaction call.
// A frontend may request acknowledges for each event, e.g. while typing in a text field, so this premature optimization is likely a win.
export class ConfigurationDefined {
    private _type : EventType;
    private _applicationName : string;
    private _availableLocales : string[];
    private _activeLocale : string;
    private _themes : Themes;
    private _resources : Resources;
    private _r : RequestId;
    get type(): EventType{
        return this._type;
    }
    set type(value: EventType){
        this._type = value;
    }
    get applicationName(): string{
        return this._applicationName;
    }
    set applicationName(value: string){
        this._applicationName = value;
    }
    get availableLocales(): string[]{
        return this._availableLocales;
    }
    set availableLocales(value: string[]){
        this._availableLocales = value;
    }
    get activeLocale(): string{
        return this._activeLocale;
    }
    set activeLocale(value: string){
        this._activeLocale = value;
    }
    get themes(): Themes{
        return this._themes;
    }
    set themes(value: Themes){
        this._themes = value;
    }
    get resources(): Resources{
        return this._resources;
    }
    set resources(value: Resources){
        this._resources = value;
    }
    get requestId(): RequestId{
        return this._r;
    }
    set requestId(value: RequestId){
        this._r = value;
    }
}

