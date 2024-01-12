import type {User} from 'oidc-client-ts';
import {UserManager} from 'oidc-client-ts';
import {defineStore} from 'pinia';
import type {Ref} from 'vue';
import {ref} from 'vue';

/**
 * After logging in, we will redirect to the URL that is in localstorage under this key.
 * This value is set during signIn and is used to redirect to a desired page.
 */
const REDIRECT_AFTER_LOGIN_STORAGE_KEY = 'auth_redirect_after_login';

/**
 * This is a definition for the state created by {@link useAuth}.
 */
export interface AuthStoreState {
    /**
     * Request the user to sign in. This will trigger a redirect to the IDP, followed by a redirect to our website after the user signed in.
     * @param redirectAfterLogin URL to redirect to after signing in. Defaults to the current location.
     *                           This must not be confused with the "redirect_uri" used in the OAuth process, which should point to the page for exchanging tokens.
     */
    signIn: (redirectAfterLogin?: string) => Promise<void>;

    /**
     * signInCallback should be called after we landed back on our website after the IDP handled a login.
     * This will perform a token exchange and will then redirect according to the preceding call to {@link signIn}.
     */
    signInCallback: () => Promise<void>;

    /**
     * Return the currently signed-in user, or null if the user is not signed in.
     * Consider using {@link user} if you need a reactive value.
     */
    getUser: () => Promise<User | null>;

    /**
     * Reactive value that contains the currently signed-in user or null if the user is not signed in.
     * Consider using {@link getUser} if you need to await the value.
     */
    user: Ref<User | null>;

    /**
     * Trigger a sign-out with a redirect to the configured post_logout_redirect_uri.
     */
    signOut: () => Promise<void>;

    init: (manager: UserManager) => void;
}

export interface UserCallback {
    (user: User | null): void;
}

export const UserChangedCallbacks = new Array<UserCallback>()

/**
 * Create a store for managing authentication.
 * See {@link AuthStoreState} for operations you can do with it.
 */
export const useAuth = defineStore<string, AuthStoreState>('authentication', () => {
    /*const userManager = new UserManager({
        authority: 'http://localhost:8080/realms/master',
        client_id: 'testclientid',
        redirect_uri: 'http://localhost:8090/oauth',
        post_logout_redirect_uri: 'http://localhost:8090',
    });*/
    let userManager: UserManager | null = null;

    // Reactive value of the currently signed-in user.
    const user: Ref<User | null> = ref(null);


    // Now define the functions needed to build the AuthStoreState.

    async function signIn(redirectAfterSignin?: string) {
        if (userManager == null) {
            return
        }

        // Store a URL to redirect to after signing in. This will be read in the signInCallback.
        const state = redirectAfterSignin || window.location.href;
        localStorage.setItem(REDIRECT_AFTER_LOGIN_STORAGE_KEY, state);

        await userManager.signinRedirect();
        console.log("signinRedirect complete", state)
    }

    async function signOut() {
        if (userManager == null) {
            return
        }

        await userManager.signoutRedirect();
    }

    async function signInCallback() {
        if (userManager == null) {
            return null
        }

        // Handle token exchange
        await userManager.signinCallback();

        // Restore the URL stored during signIn
        const redirectTo = localStorage.getItem(REDIRECT_AFTER_LOGIN_STORAGE_KEY);
        localStorage.removeItem(REDIRECT_AFTER_LOGIN_STORAGE_KEY);
        window.location.href = redirectTo || '/';
        console.log("signinCallback", redirectTo)
    }

    async function getUser(): Promise<User | null> {
        if (userManager == null) {
            console.log("auth.ts: user manager is null")
            return null
        }

        const user = await userManager.getUser();
        if (user?.expired) {
            console.log("UserManager: wtf: got an expired user!?")
        }

        return user
    }

    function init(manager: UserManager): void {
        userManager = manager

        // getUser will return the stored user. We will load the value into this store here.
        userManager.getUser().then((u) => {
            if (u?.expired) {
                user.value = null;
                console.log("userManager: user expired")
            } else {
                user.value = u;
                console.log("userManager: user updated")
            }
        });

        // Add some event listeners for when the user signed in/out to update the reactive value.
        userManager.events.addUserLoaded((u) => {
            console.log("userManager: got event that user has loaded:", u)
            user.value = u;
            UserChangedCallbacks.forEach(value => value(u))
        });

        userManager.events.addUserSignedIn(async () => {
            if (userManager == null) {
                return
            }
            console.log("userManager: user signed in")
            user.value = await userManager.getUser();
        });
        userManager.events.addUserUnloaded(() => {
            user.value = null;
            console.log("userManager: unloaded")
        });
        userManager.events.addUserSignedOut(() => {
            console.log("userManager: user signed out")
            user.value = null;
            UserChangedCallbacks.forEach(value => value(null))
        });
        userManager.events.addAccessTokenExpired(() => {
            console.log("userManager: got event that user has expired")
            user.value = null;
            UserChangedCallbacks.forEach(value => value(null))
        });
    }

    return {
        signIn,
        signInCallback,
        signOut,
        getUser,
        init,
        user,
    };
});
