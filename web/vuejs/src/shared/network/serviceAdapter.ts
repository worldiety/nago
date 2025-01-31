import {NagoEvent} from '@/shared/proto/nprotoc_gen';
import {Acknowledged} from '@/shared/protocol/ora/acknowledged';
import type {ComponentFactoryId} from '@/shared/protocol/ora/componentFactoryId';
import {ComponentInvalidated} from '@/shared/protocol/ora/componentInvalidated';
import {ConfigurationDefined} from '@/shared/protocol/ora/configurationDefined';
import type {Property} from '@/shared/protocol/ora/property';
import type {Ptr} from '@/shared/protocol/ora/ptr';
import {ScopeID} from '@/shared/protocol/ora/scopeID';
import {WindowInfo} from '@/shared/protocol/ora/windowInfo';

/**
 * Channel defines how a concrete implementation of Nago communication channel should behave.
 */
export interface Channel {
	/**
	 * sendEvent marshals the given NagoEvent and sends it over the wire to the backend.
	 * This may result in none, one or multiple follow-up events.
	 * Thus, there is no realistic correlation between a 1:1 request-response cycle and we cannot support
	 * a promise-based contract. For example, a state change may cause no invalidation, an invalidation and an error
	 * or a normal invalidation or redirect with a suppressed invalidation etc.
	 * @param evt
	 */
	sendEvent(evt: NagoEvent): void;
}

export default interface ServiceAdapter extends Channel {
	initialize(): Promise<void>;

	teardown(): Promise<void>;

	/**
	 * @deprecated
	 */
	executeFunctions(...functions: Ptr[]): Promise<ComponentInvalidated>;

	/**
	 * @deprecated
	 */
	setProperties(...properties: Property<unknown>[]): Promise<ComponentInvalidated>;

	/**
	 * @deprecated
	 */
	setPropertiesAndCallFunctions(
		properties: Property<unknown>[],
		functions: Property<Ptr>[]
	): Promise<ComponentInvalidated>;

	/**
	 * @deprecated
	 */
	createComponent(fid: ComponentFactoryId, params: Record<string, string>): Promise<ComponentInvalidated>;

	/**
	 * @deprecated
	 */
	updateWindowInfo(windowInfo: WindowInfo): void;

	/**
	 * @deprecated
	 */
	destroyComponent(ptr: Ptr): Promise<Acknowledged>;

	/**
	 * @deprecated
	 */
	getConfiguration(): Promise<ConfigurationDefined>;

	getScopeID(): ScopeID;

	/**
	 * @deprecated
	 */
	getBufferFromCache(ptr: Ptr): string | undefined;

	/**
	 * @deprecated
	 */
	setBufferToCache(ptr: Ptr, data: string): void;

	sendEvent(evt: NagoEvent): void;
}
