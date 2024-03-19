import { useAuth } from '@/stores/authStore';

export function userHeaders() {
	async function headers(): Promise<HeadersInit> {
		const auth = useAuth();
		const user = await auth.getUser;
		if (user?.expired) {
			console.log('headers: Oo user already expired and got that old one');
		}
		return {
			Authorization: `Bearer ${user?.access_token}`,
		};
	}

	return {
		headers,
	};
}

/**
 * Simple hook for making requests.
 * Automatically asks users to sign in if a 401 is received.
 */

export function useHttp() {
	const auth = useAuth();

	/**
	 * Make an HTTP request.
	 * @param url The URL to send the request to.
	 * @param method The method to make the request with.
	 * @param body The body to send in the request.
	 *             "undefined" will be an empty body, everything else will be serialized to JSON.
	 */

	async function request(url: string | URL, method = 'GET', body: undefined | any = undefined) {
		const user = await auth.getUser;

		if (user?.expired) {
			console.log('request: Oo user already expired and got that old one');
		}

		let bodyData = undefined;
		if (body !== undefined) {
			bodyData = JSON.stringify(body);
		}

		const response = await fetch(url, {
			method,
			body: bodyData,
			headers: {
				Authorization: `Bearer ${user?.access_token}`,
			},
		});

		const authRequired = response.status === 401;
		if (authRequired) {
			await auth.signIn();
		}

		if (!response.ok) {
			throw response;
		}

		try {
			return await response.clone().json(); // bei Promise als return type immer await voranstellen, sonst läuft das Programm mit dem Fehler durch
		} catch (e) {
			// TODO: hier mit dem CustomError abfangen, dass kein gültiges JSON zurückgekommen ist
			console.log('Kein gültiges JSON bekommen');
			throw e;
		}
	}

	return {
		request,
	};
}
