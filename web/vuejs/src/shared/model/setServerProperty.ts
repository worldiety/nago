export interface SetServerProperty {
	type: 'setProp'|'callFn';
	id: number;
	value: any;
}
