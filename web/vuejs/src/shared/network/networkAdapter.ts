import type { Property } from '@/shared/protocol/property';
import type Future from '@/shared/network/future';
import type { Event } from '@/shared/protocol/gen/event';
import type { EventsAggregated } from '@/shared/protocol/gen/eventsAggregated';
import { Pointer } from '@/shared/protocol/pointer';
import { SetPropertyValueRequested } from '@/shared/protocol/gen/setPropertyValueRequested';
import { FunctionCallRequested } from '@/shared/protocol/gen/functionCallRequested';
import { ComponentInvalidated } from '@/shared/protocol/gen/componentInvalidated';
import { ColorScheme } from '@/shared/protocol/colorScheme';
import { ConfigurationDefined } from '@/shared/protocol/gen/configurationDefined';
import { Acknowledged } from '@/shared/protocol/gen/acknowledged';
import { ConfigurationRequested } from '@/shared/protocol/gen/configurationRequested';

export default abstract class NetworkAdapter {

	protected pendingFutures: Map<number, Future>;
	private reqCounter: number;

	constructor() {
		this.pendingFutures = new Map();
		this.reqCounter = 0;
	}

	abstract initialize(): Promise<void>;
	abstract teardown(): Promise<void>;
	abstract executeFunctions(functions: Property<Pointer>[]): Promise<ComponentInvalidated | void>;
	abstract setProperties<T>(properties: Property<T>[]): Promise<ComponentInvalidated | void>;
	abstract setPropertiesAndCallFunctions<T>(properties: Property<T>[], functions: Property<Pointer>[]): Promise<ComponentInvalidated | void>;
	abstract createComponent(): Promise<ComponentInvalidated>;
	abstract destroyComponent(pointer: Pointer): Promise<Acknowledged>;
	abstract getConfiguration(configurationRequested: ConfigurationRequested): Promise<ConfigurationDefined>;

	protected addFuture(future: Future): void {
		// Allow a maximum of 10000 pending futures
		if (this.pendingFutures.size >= 10000) {

			const sortedPendingRequests = [...this.pendingFutures.entries()].sort(comparePendingFutures);
			this.pendingFutures.delete(sortedPendingRequests[0][0]);
		}

		this.pendingFutures.set(future.getRequestId(), future);

		function comparePendingFutures(a: [number, Future], b: [number, Future]): number {
			if (a[1].getRequestId() > b[1].getRequestId()) {
				return 1;
			} else if (a[1].getRequestId() < b[1].getRequestId()) {
				return -1;
			}
			return 0;
		}
	}

	protected resolveFuture(requestId: number): void {
		const future = this.pendingFutures.get(requestId);
		if (!future) {
			console.log(`error: got network response with unmatched requestId=${requestId}`)
		} else {
			this.pendingFutures.delete(requestId)
			future.resolveFuture(responseParsed);
		}
	}

	protected handleUnrequestedMessage(event: Event): void {

	}

	protected nextReqId(): number {
		this.reqCounter++;
		return this.reqCounter;
	}
}
