import WebSocketAdapter from '@/shared/network/webSocketAdapter';
import { defineStore } from 'pinia';
import NetworkProtocol from '@/shared/network/networkProtocol';
import type { Property } from '@/shared/model/property';
import type { PropertyFunc } from '@/shared/model/propertyFunc';

interface NetworkStoreState {
	networkProtocol: NetworkProtocol;
}

export const useNetworkStore = defineStore('networkStore', {
	state: (): NetworkStoreState => ({
		networkProtocol: new NetworkProtocol(new WebSocketAdapter()),
	}),
	actions: {
		initialize(): void {
			this.networkProtocol.initialize();
		},
		teardown(): void {
			this.networkProtocol.teardown();
		},
		invokeFunctions(...functions: PropertyFunc[]): void {
			this.networkProtocol.callFunctions(...functions);
		},
		invokeSetProperties(...properties: Property[]): void {
			this.networkProtocol.setProperties(...properties);
		},
		invokeFunctionsAndSetProperties(properties: Property[], functions: PropertyFunc[]): void {
			this.networkProtocol.setPropertiesAndCallFunctions(properties, functions);
		},
	},
});
