import { NagoEvent } from '@/shared/proto/nprotoc_gen';

/**
 * Channel defines how a concrete implementation of Nago communication channel should behave.
 */
export interface Channel {
	/**
	 * sendEvent marshals the given NagoEvent and sends it over the wire to the backend.
	 * This may result in none, one or multiple follow-up events.
	 * Thus, there is no realistic correlation between a 1:1 request-response cycle and we cannot support
	 * a promise-based contract. For example, a state change may cause no invalidation, an invalidation and an error
	 * or a normal invalidation or redirect with a suppressed invalidation etc.
	 * @param evt
	 */
	sendEvent(evt: NagoEvent): void;
}

export default interface ServiceAdapter extends Channel {
	initialize(): Promise<void>;

	teardown(): Promise<void>;

	sendEvent(evt: NagoEvent): void;
}
