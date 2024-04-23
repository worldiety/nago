import type NetworkAdapter from '@/shared/network/networkAdapter';
import {ConfigurationRequested} from "@/shared/protocol/gen/configurationRequested";
import {ColorScheme} from "@/shared/protocol/colorScheme";
import {ConfigurationDefined} from "@/shared/protocol/gen/configurationDefined";
import {ComponentFactoryId} from "@/shared/protocol/componentFactoryId";
import {ComponentInvalidated} from "@/shared/protocol/gen/componentInvalidated";
import {NewComponentRequested} from "@/shared/protocol/gen/newComponentRequested";
import {Property} from "@/shared/protocol/property";
import {Pointer} from "@/shared/protocol/pointer";
import {EventsAggregated} from "@/shared/protocol/gen/eventsAggregated";
import {SetPropertyValueRequested} from "@/shared/protocol/gen/setPropertyValueRequested";
import {FunctionCallRequested} from "@/shared/protocol/gen/functionCallRequested";
import {Event} from "@/shared/protocol/gen/event";
import {Acknowledged} from "@/shared/protocol/gen/acknowledged";
import {ComponentDestructionRequested} from "@/shared/protocol/gen/componentDestructionRequested";

export default class NetworkProtocol {

	private networkAdapter: NetworkAdapter;
	private reqCounter: number;
	private activeLocale: string;
	private unprocessedEventSubscribers: ((evt: Event) => void)[]

	constructor(networkAdapter: NetworkAdapter) {
		this.networkAdapter = networkAdapter;
		this.reqCounter = 1;
		this.activeLocale = '';
		this.unprocessedEventSubscribers = [];
	}

	async initialize(): Promise<void> {
		return this.networkAdapter.initialize();
	}

	private nextReqId(): number {
		this.reqCounter++;
		return this.reqCounter;
	}

	addUnprocessedEventSubscriber(fn: ((evt: Event) => void)) {
		this.unprocessedEventSubscribers.push(fn)
	}

	removeUnprocessedEventSubscriber(fn: ((evt: Event) => void)) {
		this.unprocessedEventSubscribers = this.unprocessedEventSubscribers.filter(obj => obj !== fn)
	}

	async destroyComponent(ptr: Pointer): Promise<Acknowledged> {
		const evt: ComponentDestructionRequested = {
			type: 'ComponentDestructionRequested',
			requestId: this.nextReqId(),
			ptr: ptr,
		};

		return this.publishToAdapter(evt.requestId, evt).then(value => {
			let evt = value as Acknowledged
			return evt
		});
	}

	async getConfiguration(colorScheme: ColorScheme, acceptLanguages: string): Promise<ConfigurationDefined> {
		const configurationRequested: ConfigurationRequested = {
			type: 'ConfigurationRequested',
			requestId: this.nextReqId(),
			acceptLanguage: acceptLanguages,
			colorScheme: colorScheme,
		};
		return this.networkAdapter.getConfiguration(configurationRequested);
	}

	async newComponent(fid: ComponentFactoryId, params: Record<string, string>): Promise<ComponentInvalidated> {
		if (this.activeLocale == "") {
			console.log("there is no configured active locale. Invoke getConfiguration to set it.")
		}

		const evt: NewComponentRequested = {
			type: 'NewComponentRequested',
			requestId: this.nextReqId(),
			activeLocale: this.activeLocale,
			factory: fid,
			values: params,
		};

		return this.publishToAdapter(evt.requestId, evt).then(value => value as ComponentInvalidated)
	}


	teardown(): void {
		this.networkAdapter.teardown();
	}

	// todo I don't believe a void choice type is a good idea...
	async callFunctions(...functions: Property<Pointer>[]): Promise<ComponentInvalidated | void> {
		return this.networkAdapter.executeFunctions(functions);
	}

	// todo I don't believe a void choice type is a good idea...
	async setProperties(...properties: Property<unknown>[]): Promise<ComponentInvalidated | void> {
		return this.networkAdapter.setProperties(properties);
	}

	// todo I don't believe a void choice type is a good idea...
	async setPropertiesAndCallFunctions(properties: Property<unknown>[], functions: Property<Pointer>[]): Promise<ComponentInvalidated | void> {
		return this.networkAdapter.setPropertiesAndCallFunctions(properties, functions);
	}
}
