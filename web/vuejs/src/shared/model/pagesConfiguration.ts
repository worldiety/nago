import type { PageConfiguration } from '@/shared/model/pageConfiguration';
import type { OIDCProvider } from '@/shared/model/oidcProvider';
import type { LivePageConfiguration } from '@/shared/model/livePageConfiguration';

export interface PagesConfiguration {
	name: string;
	pages: PageConfiguration[];
	index: string;
	oidc: OIDCProvider[]
	livePages: LivePageConfiguration[]
}
