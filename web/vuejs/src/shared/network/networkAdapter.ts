import type { Property } from '@/shared/protocol/property';
import type { Event } from '@/shared/protocol/gen/event';
import type { Pointer } from '@/shared/protocol/pointer';
import type { ConfigurationDefined } from '@/shared/protocol/gen/configurationDefined';
import type { Acknowledged } from '@/shared/protocol/gen/acknowledged';
import type { ComponentInvalidated } from '@/shared/protocol/gen/componentInvalidated';
import type { ColorScheme } from '@/shared/protocol/colorScheme';
import type { ComponentFactoryId } from '@/shared/protocol/componentFactoryId';

export default abstract class NetworkAdapter {

	protected activeLocale: string;
	private requestId: number;
	private unrequestedEventSubscribers: ((evt: Event) => void)[];

	constructor() {
		this.activeLocale = '';
		this.requestId = 0;
		this.unrequestedEventSubscribers = [];
	}

	abstract initialize(): Promise<void>;
	abstract teardown(): Promise<void>;
	abstract executeFunctions(functions: Property<Pointer>[]): Promise<ComponentInvalidated>;
	abstract setProperties<T>(properties: Property<T>[]): Promise<ComponentInvalidated>;
	abstract setPropertiesAndCallFunctions<T>(properties: Property<T>[], functions: Property<Pointer>[]): Promise<ComponentInvalidated>;
	abstract createComponent(fid: ComponentFactoryId, params: Record<string, string>): Promise<ComponentInvalidated>;
	abstract destroyComponent(ptr: Pointer): Promise<Acknowledged>;
	abstract getConfiguration(colorScheme: ColorScheme, acceptLanguages: string): Promise<ConfigurationDefined>;

	addUnrequestedEventSubscriber(fn: ((event: Event) => void)) {
		this.unrequestedEventSubscribers.push(fn);
	}

	removeUnrequestedEventSubscriber(fn: ((evt: Event) => void)) {
		this.unrequestedEventSubscribers = this.unrequestedEventSubscribers.filter(obj => obj !== fn)
	}

	protected handleUnrequestedEvent(event: Event): void {
		this.unrequestedEventSubscribers.forEach(fn => fn(event));
	}

	protected nextRequestId(): number {
		this.requestId++;
		return this.requestId;
	}
}
