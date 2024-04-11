import WebSocketAdapter from '@/shared/network/webSocketAdapter';
import { defineStore } from 'pinia';
import NetworkProtocol from '@/shared/network/networkProtocol';
import type { Property } from '@/shared/model/property';
import type { PropertyFunc } from '@/shared/model/propertyFunc';
import type { Invalidation } from '@/shared/model/invalidation';

interface NetworkStoreState {
	networkProtocol: NetworkProtocol;
}

export const useNetworkStore = defineStore('networkStore', {
	state: (): NetworkStoreState => ({
		networkProtocol: new NetworkProtocol(new WebSocketAdapter()),
	}),
	actions: {
		async initialize(): Promise<Invalidation> {
			return this.networkProtocol.initialize();
		},
		teardown(): void {
			this.networkProtocol.teardown();
		},
		async invokeFunctions(...functions: PropertyFunc[]): Promise<Invalidation|void> {
			return this.networkProtocol.callFunctions(...functions);
		},
		async invokeSetProperties(...properties: Property[]): Promise<Invalidation|void> {
			return this.networkProtocol.setProperties(...properties);
		},
		async invokeFunctionsAndSetProperties(properties: Property[], functions: PropertyFunc[]): Promise<Invalidation|void> {
			return this.networkProtocol.setPropertiesAndCallFunctions(properties, functions);
		},
	},
});
