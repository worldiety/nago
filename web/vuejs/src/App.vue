<script setup lang="ts">
import { RouterView, useRoute, useRouter } from 'vue-router';
import Page from '@/views/Page.vue';
import { ref } from 'vue';
import { useAuth } from '@/stores/authStore';
import { UserManager } from 'oidc-client-ts';
import { fetchApplication } from '@/api/application/appRepository';
import UiErrorMessage from '@/components/UiErrorMessage.vue';
import { ApplicationError, type CustomError, useErrorHandling } from '@/composables/errorhandling';
import i18n from '@/i18n';
import type { PagesConfiguration } from '@/shared/model/pagesConfiguration';

enum State {
	LoadingRoutes,
	ShowRoutes,
	Error,
}

const errorHandler = useErrorHandling();
const router = useRouter();
const route = useRoute();
const auth = useAuth();
const state = ref(State.LoadingRoutes);

//TODO: Torben baut zukünftig /health ein, der einen 200er und eine json-response zurückgibt, wenn der Service grundsätzlich läuft

async function init(): Promise<void> {
	let anchor: string;

	try {
		const app = await fetchApplication();

		if (app.oidc?.length > 0) {
			/*auth.init(new UserManager({
        authority: 'http://localhost:8080/realms/master',
        client_id: 'testclientid',
        redirect_uri: 'http://localhost:8090/oauth',
        post_logout_redirect_uri: 'http://localhost:8090',
      }))*/
			const provider = app.oidc.at(0);
			if (provider) {
				auth.init(
					new UserManager({
						authority: provider.authority,
						client_id: provider.clientID,
						redirect_uri: provider.redirectURL,
						post_logout_redirect_uri: provider.postLogoutRedirectUri,
					})
				);
			}
		}

		app.livePages.forEach((page) => {
			anchor = page.anchor.replaceAll('{', ':');
			anchor = anchor.replaceAll('}', '?');
			anchor = anchor.replaceAll('-', '\\-'); //OMG regex
			router.addRoute({ path: anchor, component: Page, meta: { page } });
			console.log('registered route', anchor);
		});

		// Update router with current route, to load the dynamically configured page.
		await router.replace(route);

		state.value = State.ShowRoutes;

		if (router.currentRoute.value.path === '/' && app.index != null && app.index != '') {
			console.log('app requires index rewrite to ', app.index);
			router.replace(app.index);
		}
	} catch (e: ApplicationError) {
		errorHandler.handleError(e);
	}
}

init();
</script>

<template>
	<div v-if="errorHandler.error.value" class="flex h-screen items-center justify-center">
		<UiErrorMessage :error="errorHandler.error.value"> </UiErrorMessage>
	</div>

	<RouterView v-if="state === State.ShowRoutes" />
</template>
