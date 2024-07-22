/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */


import type { EventType } from '@/shared/protocol/ora/eventType';
import type { RequestId } from '@/shared/protocol/ora/requestId';
import type { Themes } from '@/shared/protocol/ora/themes';

/**
 * A ConfigurationDefined event is the response to a [ConfigurationRequested] event.
 * According to the locale request, string and svg resources can be localized by the backend.
 * The returned locale is the actually picked locale from the requested locale query string.
 * 
 * It looks quite obfuscated, however this minified version is intentional, because it may succeed each transaction call.
 * A frontend may request acknowledges for each event, e.g. while typing in a text field, so this premature optimization is likely a win.
 */
export interface ConfigurationDefined {
    // Type
    type: 'ConfigurationDefined'/*EventType*/;
    // ApplicationID
    applicationID: string;
    // ApplicationName
    applicationName: string;
    // ApplicationVersion
    applicationVersion: string;
    // AvailableLocales
    availableLocales: string[];
    // ActiveLocale
    activeLocale: string;
    // Themes
    themes: Themes;
    // RequestId
    r /*RequestId*/: RequestId;
}

