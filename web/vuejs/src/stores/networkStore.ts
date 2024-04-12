import WebSocketAdapter from '@/shared/network/webSocketAdapter';
import { defineStore } from 'pinia';
import NetworkProtocol from '@/shared/network/networkProtocol';
import type { Property } from '@/shared/model/property';
import type { PropertyFunc } from '@/shared/model/propertyFunc';
import type { Invalidation } from '@/shared/model/invalidation';
import {ConfigurationDefined} from "@/shared/protocol/gen/configurationDefined";
import {ColorScheme} from "@/shared/protocol/colorScheme";
import {ComponentFactoryId} from "@/shared/protocol/componentFactoryId";
import {ComponentInvalidated} from "@/shared/protocol/gen/componentInvalidated";

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
