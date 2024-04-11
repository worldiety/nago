export default interface NetworkAdapter {
	initialize(): Promise<void>;
	teardown(): void;
	publish(payloadRaw: string): void;
	subscribe(resolve: (responseRaw: string) => void): void;
}
