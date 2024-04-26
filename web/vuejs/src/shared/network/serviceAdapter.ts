import type { Property } from '@/shared/protocol/property';
import type { Pointer } from '@/shared/protocol/pointer';
import type { ColorScheme } from '@/shared/protocol/colorScheme';
import type { ComponentFactoryId } from '@/shared/protocol/componentFactoryId';
import { ComponentInvalidated } from '@/shared/protocol/gen/componentInvalidated';
import { Acknowledged } from '@/shared/protocol/gen/acknowledged';
import { ConfigurationDefined } from '@/shared/protocol/gen/configurationDefined';

export default interface ServiceAdapter {

	initialize(): Promise<void>;
	teardown(): Promise<void>;
	executeFunctions(...functions: Property<Pointer>[]): Promise<ComponentInvalidated>;
	setProperties(...properties: Property<unknown>[]): Promise<ComponentInvalidated>;
	setPropertiesAndCallFunctions(properties: Property<unknown>[], functions: Property<Pointer>[]): Promise<ComponentInvalidated>;
	createComponent(fid: ComponentFactoryId, params: Record<string, string>): Promise<ComponentInvalidated>;
	destroyComponent(ptr: Pointer): Promise<Acknowledged>;
	getConfiguration(colorScheme: ColorScheme, acceptLanguages: string): Promise<ConfigurationDefined>;
}
