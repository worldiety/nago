export interface PagesConfiguration {
	name: string;
	pages: PageConfiguration[];
	index: string;
	oidc: OIDCProvider[]
	livePages: LivePageConfiguration[]
}
