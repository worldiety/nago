import type { LivePageConfiguration } from '@/shared/model/livePageConfiguration';
import type { OIDCProvider } from '@/shared/model/oidcProvider';
import type { PageConfiguration } from '@/shared/model/pageConfiguration';

export interface PagesConfiguration {
	name: string;
	pages: PageConfiguration[];
	index: string;
	oidc: OIDCProvider[];
	livePages: LivePageConfiguration[];
}
