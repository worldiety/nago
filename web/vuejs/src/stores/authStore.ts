import type {User} from 'oidc-client-ts';
import type {UserManager} from 'oidc-client-ts';
import {defineStore} from 'pinia';

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
     * Reactive value that contains the currently signed-in user or null if the user is not signed in.
     * Consider using {@link getUser} if you need to await the value.
     */
    user: User | null;
		userManager: UserManager | null;
}

export interface UserCallback {
    (user: User | null): void;
}

export const UserChangedCallbacks = new Array<UserCallback>()

/**
 * Create a store for managing authentication.
 * See {@link AuthStoreState} for operations you can do with it.
 */
export const useAuth = defineStore('authentication', {
	state: (): AuthStoreState => ({
		userManager: null,
		user: null,
	}),
	getters: {
		/**
		 * Return the currently signed-in user, or null if the user is not signed in.
		 * Consider using {@link user} if you need a reactive value.
		 */
		getUser: async (state): Promise<User | null> => {
			if (state.userManager == null) {
				console.log("auth.ts: user manager is null")
				return null
			}

			const user = await state.userManager.getUser();
			if (user?.expired) {
				console.log("UserManager: wtf: got an expired user!?")
			}

			return user
		},
	},
	actions: {
		/**
		 * Request the user to sign in. This will trigger a redirect to the IDP, followed by a redirect to our website after the user signed in.
		 * @param redirectAfterLogin URL to redirect to after signing in. Defaults to the current location.
		 *                           This must not be confused with the "redirect_uri" used in the OAuth process, which should point to the page for exchanging tokens.
		 */
		async signIn(redirectAfterSignin?: string): Promise<void> {
			if (this.userManager == null) {
				return
			}

			// Store a URL to redirect to after signing in. This will be read in the signInCallback.
			const state = redirectAfterSignin || window.location.href;
			localStorage.setItem(REDIRECT_AFTER_LOGIN_STORAGE_KEY, state);

			await this.userManager.signinRedirect();
			console.log("signinRedirect complete", state)
		},
		/**
		 * Trigger a sign-out with a redirect to the configured post_logout_redirect_uri.
		 */
		async signOut(): Promise<void> {
			if (this.userManager == null) {
				return
			}

			await this.userManager.signoutRedirect();
		},
		/**
		 * signInCallback should be called after we landed back on our website after the IDP handled a login.
		 * This will perform a token exchange and will then redirect according to the preceding call to {@link signIn}.
		 */
		async signInCallback() {
			if (this.userManager == null) {
				return null
			}

			// Handle token exchange
			await this.userManager.signinCallback();

			// Restore the URL stored during signIn
			const redirectTo = localStorage.getItem(REDIRECT_AFTER_LOGIN_STORAGE_KEY);
			localStorage.removeItem(REDIRECT_AFTER_LOGIN_STORAGE_KEY);
			window.location.href = redirectTo || '/';
			console.log("signinCallback", redirectTo)
		},
		init(manager: UserManager): void {
			this.userManager = manager

			// getUser will return the stored user. We will load the value into this store here.
			this.userManager.getUser().then((u) => {
				if (u?.expired) {
					this.user = null;
					console.log("userManager: user expired")
				} else {
					this.user = u;
					console.log("userManager: user updated")
				}
			});

			// Add some event listeners for when the user signed in/out to update the reactive value.
			this.userManager.events.addUserLoaded((u) => {
				console.log("userManager: got event that user has loaded:", u)
				this.user = u;
				UserChangedCallbacks.forEach(value => value(u))
			});

			this.userManager.events.addUserSignedIn(async () => {
				if (this.userManager == null) {
					return
				}
				console.log("userManager: user signed in")
				this.user = await this.userManager.getUser();
			});
			this.userManager.events.addUserUnloaded(() => {
				this.user = null;
				console.log("userManager: unloaded")
			});
			this.userManager.events.addUserSignedOut(() => {
				console.log("userManager: user signed out")
				this.user = null;
				UserChangedCallbacks.forEach(value => value(null))
			});
			this.userManager.events.addAccessTokenExpired(() => {
				console.log("userManager: got event that user has expired")
				this.user = null;
				UserChangedCallbacks.forEach(value => value(null))
			});
		},
	},
});
