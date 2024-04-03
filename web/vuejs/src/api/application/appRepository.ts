import type { PagesConfiguration } from '@/shared/model/pagesConfiguration';
import {HttpRequest} from "@/shared/http";

//TODO: Klasse anlegen

export async function fetchApplication(): Promise<PagesConfiguration | undefined> {
		  return await HttpRequest.get(import.meta.env.VITE_HOST_BACKEND + '/api/v1/ui/application')
			.fetch()
}
