import type {Property} from '@/shared/protocol/ora/property';
import type {Ptr} from '@/shared/protocol/ora/ptr';
import type {ComponentFactoryId} from '@/shared/protocol/ora/componentFactoryId';
import {ComponentInvalidated} from '@/shared/protocol/ora/componentInvalidated';
import {Acknowledged} from '@/shared/protocol/ora/acknowledged';
import {ConfigurationDefined} from '@/shared/protocol/ora/configurationDefined';
import {ScopeID} from "@/shared/protocol/ora/scopeID";
import {WindowInfo} from "@/shared/protocol/ora/windowInfo";
import {NagoEvent} from "@/shared/proto/nprotoc_gen";

export default interface ServiceAdapter {

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
	setPropertiesAndCallFunctions(properties: Property<unknown>[], functions: Property<Ptr>[]): Promise<ComponentInvalidated>;

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
	getBufferFromCache(ptr: Ptr): string | undefined

	/**
	 * @deprecated
	 */
	setBufferToCache(ptr: Ptr, data: string): void;

	sendEvent(evt: NagoEvent): void;
}
