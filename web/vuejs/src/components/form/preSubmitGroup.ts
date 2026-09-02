export const PRE_SUBMIT_GROUP = 'preSubmitGroup';

export class PreSubmitGroup {
	private readonly memberCallbacks: (() => void)[];

	constructor() {
		this.memberCallbacks = [];
	}

	public join(callback: () => void): void {
		this.memberCallbacks.push(callback);
	}

	public leave(callback: () => void): void {
		const index = this.memberCallbacks.indexOf(callback);
		if (index !== -1) {
			this.memberCallbacks.splice(index, 1);
		}
	}

	public async execute(): Promise<void> {
		for (const callback of this.memberCallbacks) {
			callback();
		}
	}
}
