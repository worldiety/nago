import { useAuth } from "@/stores/auth";

/**
 * Simple hook for making requests.
 * Automatically asks users to sign in if a 401 is received.
 */
export function useHttp() {
    const auth = useAuth();

    async function request(url: string, method = "GET") {
        const user = await auth.getUser();

        const response = await fetch(url, {
            method,
            headers: {
                "Authorization": `Bearer ${user?.access_token}`
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