export interface Redirection {
	type: 'Redirect';
	url: string;
	direction: 'forward' | 'backward';
	redirect: boolean;
}
