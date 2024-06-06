/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */


import type { ColorScheme } from '@/shared/protocol/ora/colorScheme';
import type { EventType } from '@/shared/protocol/ora/eventType';
import type { RequestId } from '@/shared/protocol/ora/requestId';
import type { WindowInfo } from '@/shared/protocol/ora/windowInfo';

/**
 * ConfigurationRequested is issued by the frontend to get the applications general configuration.
 * A backend developer has potentially defined a lot of configuration details about the application.
 * For example, there may be a color theme, customized icons, image resources, an application name and the available set of navigations, launch intents or other meta information.
 * It is expected, that this only happens once during initialization of the frontend process.
 */
export interface ConfigurationRequested {
    type: 'ConfigurationRequested'/*EventType*/;
    acceptLanguage: string;
    colorScheme: ColorScheme;
    windowInfo: WindowInfo;
    r /*RequestId*/: RequestId;
}

