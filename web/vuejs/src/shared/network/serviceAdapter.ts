import type {Property} from '@/shared/protocol/ora/property';
import type {Ptr} from '@/shared/protocol/ora/ptr';
import type {ComponentFactoryId} from '@/shared/protocol/ora/componentFactoryId';
import {ComponentInvalidated} from '@/shared/protocol/ora/componentInvalidated';
import {Acknowledged} from '@/shared/protocol/ora/acknowledged';
import {ConfigurationDefined} from '@/shared/protocol/ora/configurationDefined';
import {ScopeID} from "@/shared/protocol/ora/scopeID";
import {WindowInfo} from "@/shared/protocol/ora/windowInfo";

export default interface ServiceAdapter {

	initialize(): Promise<void>;

	teardown(): Promise<void>;

	executeFunctions(...functions: Property<Ptr>[]): Promise<ComponentInvalidated>;

	setProperties(...properties: Property<unknown>[]): Promise<ComponentInvalidated>;

	setPropertiesAndCallFunctions(properties: Property<unknown>[], functions: Property<Ptr>[]): Promise<ComponentInvalidated>;

	createComponent(fid: ComponentFactoryId, params: Record<string, string>): Promise<ComponentInvalidated>;

	updateWindowInfo(windowInfo: WindowInfo): void;

	destroyComponent(ptr: Ptr): Promise<Acknowledged>;

	getConfiguration(): Promise<ConfigurationDefined>;

	getScopeID(): ScopeID;
}
