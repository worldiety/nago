import WebSocketAdapter from '@/shared/network/webSocketAdapter';
import { defineStore } from 'pinia';
import NetworkProtocol from '@/shared/network/networkProtocol';
import type { PropertyFunc } from '@/shared/model/propertyFunc';
import type { Invalidation } from '@/shared/model/invalidation';
import {ConfigurationDefined} from "@/shared/protocol/gen/configurationDefined";
import {ColorScheme} from "@/shared/protocol/colorScheme";
import {ComponentFactoryId} from "@/shared/protocol/componentFactoryId";
import {ComponentInvalidated} from "@/shared/protocol/gen/componentInvalidated";
import {Pointer} from "@/shared/protocol/pointer";
import {Property} from "@/shared/protocol/property";
import {Event} from "@/shared/protocol/gen/event";

interface NetworkStoreState {
	networkProtocol: NetworkProtocol;
}

export const useNetworkStore = defineStore('networkStore', {
	state: (): NetworkStoreState => ({
		networkProtocol: new NetworkProtocol(new WebSocketAdapter()),
	}),
	actions: {
		async initialize(): Promise<void> {
			return this.networkProtocol.initialize();
		},
		async getConfiguration(colorScheme:ColorScheme, acceptLanguages: string): Promise<ConfigurationDefined>{
			return this.networkProtocol.getConfiguration(colorScheme,acceptLanguages)
		},

		async newComponent(fid:ComponentFactoryId, params : Map<string,string>):Promise<ComponentInvalidated>{
			return this.networkProtocol.newComponent(fid,params)
		},

		addUnprocessedEventSubscriber(fn: ((evt: Event) => void)) {
			this.networkProtocol.addUnprocessedEventSubscriber(fn)
		},

		removeUnprocessedEventSubscriber(fn: ((evt: Event) => void)){
			this.networkProtocol.removeUnprocessedEventSubscriber(fn)
		},

		teardown(): void {
			this.networkProtocol.teardown();
		},
		async invokeFunctions(...functions: Property<Pointer>[]): Promise<Invalidation|void> {
			return this.networkProtocol.callFunctions(...functions);
		},
		async invokeSetProperties(...properties: Property<unknown>[]): Promise<Invalidation|void> {
			return this.networkProtocol.setProperties(...properties);
		},
		async invokeFunctionsAndSetProperties(properties: Property<unknown>[], functions: Property<Pointer>[]): Promise<Invalidation|void> {
			return this.networkProtocol.setPropertiesAndCallFunctions(properties, functions);
		},
	},
});
