import type { Property } from '@/shared/protocol/property';
import type { Pointer } from '@/shared/protocol/pointer';
import type { ConfigurationDefined } from '@/shared/protocol/gen/configurationDefined';
import type { Acknowledged } from '@/shared/protocol/gen/acknowledged';
import type { ComponentInvalidated } from '@/shared/protocol/gen/componentInvalidated';
import type { ColorScheme } from '@/shared/protocol/colorScheme';
import type { ComponentFactoryId } from '@/shared/protocol/componentFactoryId';

export default interface NetworkAdapter {

	initialize(): Promise<void>;
	teardown(): Promise<void>;
	executeFunctions(functions: Property<Pointer>[]): Promise<ComponentInvalidated>;
	setProperties(properties: Property<unknown>[]): Promise<ComponentInvalidated>;
	setPropertiesAndCallFunctions(properties: Property<unknown>[], functions: Property<Pointer>[]): Promise<ComponentInvalidated>;
	createComponent(fid: ComponentFactoryId, params: Record<string, string>): Promise<ComponentInvalidated>;
	destroyComponent(ptr: Pointer): Promise<Acknowledged>;
	getConfiguration(colorScheme: ColorScheme, acceptLanguages: string): Promise<ConfigurationDefined>;
}
