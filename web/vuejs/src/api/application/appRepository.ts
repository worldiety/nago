import { useHttp } from '@/shared/http';
import type { PagesConfiguration } from '@/shared/model/pagesConfiguration';

//TODO: Klasse anlegen

// TODO: hier die fetchApplication Funktion schreiben
// Funktion schreiben, die den HTTP Client verwendet
export async function fetchApplication(): Promise<PagesConfiguration> {
	const client = useHttp();
	return await client.request(import.meta.env.VITE_HOST_BACKEND + '/api/v1/ui/application');
}
