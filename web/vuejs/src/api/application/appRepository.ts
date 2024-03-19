import { useHttp } from '@/shared/http';
import type { PagesConfiguration } from '@/shared/model/pagesConfiguration';

//TODO: Klasse anlegen

export async function fetchApplication(): Promise<PagesConfiguration> {
	const client = useHttp();
	return await client.request(import.meta.env.VITE_HOST_BACKEND + '/api/v1/ui/application');
}
