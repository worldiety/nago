import { useAuth } from '@/stores/auth';

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
    async function request(url: string, method = 'GET', body: undefined | any = undefined) {
        const user = await auth.getUser();

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

        return response;
    }

    return {
        request,
    };
}
