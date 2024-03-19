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

const contentType = ref<string | null>('');

//TODO: JSON Dateien überarbeiten, sodass die Custom Codes und die HTTP-Codes nicht vermischt werden (ein Objekt dazwischen verwenden)

//TODO: i18n so überarbeiten, dass ich das Plugin verwende (vue-i18n)

//TODO: nago example erweitern um ein Formular, das ich abschicke bzw. Datei Upload, um da das Fehlerhandling z.B. Internet vorhanden testen.

//TODO: überarbeiten, dass bei jedem Request auf Fehler überprüft wird. Im Moment findet die Überprüfung nur einmal am Anfang statt.

//TODO: Torben baut zukünftig /health ein, der einen 200er und eine json-response zurückgibt, wenn der Service grundsätzlich läuft

async function init(): Promise<void> {
	let anchor: string;
	let response: Response | void | null = null;

	//TODO: Kann hier weg
	// check internet connection
	if (!navigator.onLine) {
		console.log('Keine Internetverbindung!');
	}

	try {
		//TODO: Mit Torben absprechen, ob wir uns auf axios oder fetch festlegen. Malte empfiehlt mir axios, da es erstmal einfacher zu benutzen ist

		//TODO: in eine eigene Datei und Funktion (fetchApplication) auslagern. HTTP Client bauen (siehe Maltes Beispiel).
		// Datei bauen, die einfach REST Anfragen durchführt und schaut, ob alles in Ordnung ist. Wenn nicht, Fehler auswerfen (throw)
		// UI-Komponente bauen, um den Fehler darzustellen

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

		//TODO: Hier überprüfen, ob ich einen Statuscode bekomme. Failed to fetch bedeutet fast immer, dass keine Internetverbindung vorliegt
		// Kann man aber mit navigator.online gegenprüfen
		// Auf folgende Statuscode prüfen:
		// 401 = Keine Authentifizierung vorhanden
		// 403 = Man ist angemeldet, aber was ich tun möchte, darf ich nicht
		// 404 = angefragte Entität ist nicht vorhanden (braucht man eigentlich nur, wenn man einen bestimmten Key abfragen möchte)
		// 500 = Irgendwas ist fehlgeschlagen. Unbekannt was (default Fehler)
		// je nachdem, wie man es deployed 502
		// 503
		// 504
		// Doku Statuscodes: https://developer.mozilla.org/en-US/docs/Web/HTTP/Status
		// ggfs. andere Statuscodes mit Torben besprechen (Torben sollte eine API-Dokumentation angelegt haben)
		// ggfs. einbauen, ob ich die richtige Antwort erhalten habe. Sollte eigentlich nicht nötig sein. Wenn doch, dann meistens
		// Programmierfehler

		// TODO:i18n Dateien anlegen und da die einzelnen Fehlercodes hinterlegen. Dann hier darauf zugreifen
		// neue Fehlerkomponente und Fehlerobjekt erstellen, die meine Fehlermeldung anzeigen. Diesen Bereich in meine api auslagern und
		// dort mit throw arbeiten

		console.log('Dieser Fehler ist aufgetreten: ' + rawError);
	}
}

init();
</script>

<template>
	<div></div>
	<div v-if="errorHandler.error.value" class="flex h-screen items-center justify-center">
		<UiErrorMessage :error="errorHandler.error.value"> </UiErrorMessage>
		<!--
      <div class="errorMessage">
        <p> {{errorMessage}}</p>
        <button class="border border-black p-0.5 text-sm bg-white" @click="toggleErrorInfo">Mehr Informationen</button>
        <p v-if="showAdditionalInformation">{{ additionalInformation }}</p>
      </div>
      -->
	</div>

	<RouterView v-if="state === State.ShowRoutes" />
</template>
