import WebSocketAdapter from '@/shared/network/webSocketAdapter';
import { defineStore } from 'pinia';
import type {ConfigurationDefined} from "@/shared/protocol/gen/configurationDefined";
import type {ColorScheme} from "@/shared/protocol/colorScheme";
import type {ComponentFactoryId} from "@/shared/protocol/componentFactoryId";
import type {ComponentInvalidated} from "@/shared/protocol/gen/componentInvalidated";
import type {Pointer} from "@/shared/protocol/pointer";
import type {Property} from "@/shared/protocol/property";
import type {Event} from "@/shared/protocol/gen/event";
import type {Acknowledged} from "@/shared/protocol/gen/acknowledged";
import type NetworkAdapter from '@/shared/network/networkAdapter';

interface NetworkStoreState {
	networkAdapter: NetworkAdapter;
}

export const useNetworkStore = defineStore('networkStore', {
	state: (): NetworkStoreState => ({
		networkAdapter: new WebSocketAdapter(),
	}),
	actions: {
		async initialize(): Promise<void> {
			return this.networkAdapter.initialize();
		},
		async getConfiguration(colorScheme:ColorScheme, acceptLanguages: string): Promise<ConfigurationDefined>{
			return this.networkAdapter.getConfiguration(colorScheme,acceptLanguages)
		},

		async destroyComponent(ptr :Pointer):Promise<Acknowledged>{
			return this.networkAdapter.destroyComponent(ptr)
		},

		async newComponent(fid:ComponentFactoryId, params : Record<string,string>):Promise<ComponentInvalidated>{
			return this.networkAdapter.createComponent(fid, params)
		},

		addUnrequestedEventSubscriber(fn: ((evt: Event) => void)) {
			this.networkAdapter.addUnrequestedEventSubscriber(fn)
		},

		removeUnprocessedEventSubscriber(fn: ((evt: Event) => void)){
			this.networkAdapter.removeUnrequestedEventSubscriber(fn)
		},

		async teardown(): Promise<void> {
			return this.networkAdapter.teardown();
		},
		async invokeFunctions(...functions: Property<Pointer>[]): Promise<ComponentInvalidated> {
			return this.networkAdapter.executeFunctions(functions);
		},
		async invokeSetProperties(...properties: Property<unknown>[]): Promise<ComponentInvalidated> {
			return this.networkAdapter.setProperties(properties);
		},
		async invokeFunctionsAndSetProperties(properties: Property<unknown>[], functions: Property<Pointer>[]): Promise<ComponentInvalidated> {
			return this.networkAdapter.setPropertiesAndCallFunctions(properties, functions);
		},
	},
});
