export default interface NetworkAdapter {
	initialize(): void;
	teardown(): void;
	publish(payloadRaw: string): void;
	subscribe(resolve: (responseRaw: string) => void): void;
}
