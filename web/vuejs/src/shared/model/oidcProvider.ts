// deprecated remove me
export interface OIDCProvider {
	name: string;
	authority: string;
	clientID: string;
	clientSecret: string;
	redirectURL: string;
	postLogoutRedirectUri: string;
}
