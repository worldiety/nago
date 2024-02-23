

<script setup lang="ts">
import {RouterView, useRoute, useRouter} from 'vue-router';
import Page from '@/views/Page.vue';
import {ref, onMounted} from 'vue';
import {fetchBackendData} from "@/api/application/appService";
import type {PagesConfiguration} from '@/shared/model';
import {useAuth} from "@/stores/auth";
import {UserManager} from "oidc-client-ts";
import {LiveMessage} from "@/shared/livemsg";
import axios from 'axios';
import {Store} from "pinia";

const router = useRouter();
const route = useRoute();

enum State {
  LoadingRoutes,
  ShowRoutes,
  Error,
}

const auth = useAuth();
const state = ref(State.LoadingRoutes);

//TODO: Types festlegen. Bei Variablen die null sein können, mit null oder "leer" initialisieren
//any ist immer böse
const errorMessage = ref<string>('')
const additionalInformation = ref<string>('')

const captivePortal = ref<string>('')
const captiveOk = ref<boolean>(false)
const showAdditionalInformation = ref<boolean>(false)
const contentType = ref<string | null>('')


//TODO: überarbeiten, dass bei jedem Request auf Fehler überprüft wird. Im Moment findet die Überprüfung nur einmal am
// Anfang statt

//TODO: Die ganzen checks auslagern, also die App.vue aufräumen
async function checkCaptivePortal():Promise<boolean> {
  try {

    //TODO: Statuscode checken
    //TODO: Torben anschreiben, ob es einen Statusendpunkt gibt, der grundsätzlich 200 zurückgibt, wenn er erreichbar ist
    // Torben baut zuünkftig /health ein, der 200er und eine json-response zurückgibt, wenn der Service grundsätzlich läuft
    const response = await axios.get<string>('/api/')
    captivePortal.value = response.data
    console.log('Captive Portal Check: ' + response.status)

    return captiveOk.value = true

  } catch (error) {
    errorMessage.value = "Keine Verbindung möglich. Bitte Rechte prüfen."
    additionalInformation.value = "Captive Portal Check fehlgeschlagen."
    state.value = State.Error
    return captiveOk.value = false

  }
}

async function checkInternetConnection():Promise<boolean> {
  if (!navigator.onLine) {
    errorMessage.value = "Keine Internetverbindung vorhanden. Bitte Verbindung überprüfen.";
    additionalInformation.value = "Router und Kabel überprüfen. Eventuell WLAN-Verbindung wiederherstellen.";
    state.value = State.Error
    return false;
  }
  return true;
}

async function toggleInfo():Promise<void> {
  showAdditionalInformation.value = !showAdditionalInformation.value
}



async function init():Promise<void> {
  let dataFromBackend: string[] | null = null

  onMounted(async () => {
    try {
      dataFromBackend = await fetchBackendData()
      console.log('Data from Backend: ' + dataFromBackend)
    } catch (error) {
      console.log('Fehler!')
    }
  })

  if (!await checkInternetConnection()) {
    return
  }

  /*
  captiveOk.value = await checkCaptivePortal()

  if (!captiveOk.value) {
      return;
  }

   */




  let response: Response | null = null

  try {
    //TODO: Mit Torben absprechen, ob wir uns auf axios oder fetch festlegen. Malte empfiehlt mir axios, da es erstmal einfacher zu benutzen ist
    response = await fetch(import.meta.env.VITE_HOST_BACKEND + '/api/v1/ui/application');
    contentType.value = response.headers.get('Content-Type');
    console.log('Weiterer Captive Portal Check: ' + response.status)

    const app: PagesConfiguration = await response.json();

    if (app.oidc?.length>0){
      /*auth.init(new UserManager({
        authority: 'http://localhost:8080/realms/master',
        client_id: 'testclientid',
        redirect_uri: 'http://localhost:8090/oauth',
        post_logout_redirect_uri: 'http://localhost:8090',
      }))*/
      let provider = app.oidc.at(0)
      auth.init(new UserManager({
        authority: provider.authority,
        client_id: provider.clientID,
        redirect_uri: provider.redirectURL,
        post_logout_redirect_uri: provider.postLogoutRedirectUri,
      }))
    }



    app.livePages.forEach((page) => {
      let anchor = page.anchor.replaceAll("{", ":")
      anchor = anchor.replaceAll("}", "?")
      anchor = anchor.replaceAll("-", "\\-") //OMG regex
      router.addRoute({path: anchor, component: Page, meta: {page}});
      console.log('registered route', anchor);
    });

    // Update router with current route, to load the dynamically configured page.
    await router.replace(route)




    state.value = State.ShowRoutes;


    if (router.currentRoute.value.path==="/" && app.index != null && app.index != "") {
      console.log("app requires index rewrite to ", app.index)
      router.replace(app.index)
    }



  } catch (e) {
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

    // TOOO:i18n Dateien anlegen und da die einzelnen Fehlercodes hinterlegen. Dann hier darauf zugreifen
    // neue Fehlerkomponente und Fehlerobjekt erstellen, die meine Fehlermeldung anzeigen. Diesen Bereich in meine api auslagern und
    // dort mit throw arbeiten

    state.value = State.Error

    console.log('Dieser Fehler ist aufgetreten: ' + e)


    switch(response?.status) {
      case 401: {
        errorMessage.value = 'Authentifizierung fehlgeschlagen. Bitte einloggen.'
        additionalInformation.value = 'HTTP-Statuscode: ' + response.status
        break;
      }
      case 403: {
        errorMessage.value = 'Autorisierung fehlgeschlagen. Bitte Rechte prüfen.'
        additionalInformation.value = 'HTTP-Statuscode: ' + response.status
        break;
      }
      case 404: {
        errorMessage.value = 'Seite nicht gefunden. Bitte URL überprüfen.'
        additionalInformation.value = 'HTTP-Statuscode: ' + response.status
        break;
      }
      case 502: {
        errorMessage.value = 'Ungültige Antwort erhalten. Bitte später erneut versuchen.'
        additionalInformation.value = 'HTTP-Statuscode: ' + response.status
        break;
      }
      case 503: {
        errorMessage.value = 'Die Anfrage konnte nicht verarbeitet werden. Bitte später erneut versuchen.'
        additionalInformation.value = 'HTTP-Statuscode: ' + response.status
        break;
      }
      case 504: {
        errorMessage.value = 'Zeitüberschreitung. Bitte später erneut versuchen.'
        additionalInformation.value = 'HTTP-Statuscode: ' + response.status
        break;
      }

      //TODO: auf 500 und auf undefined abfragen (undefined höchstwahrscheinlich Netzwerkfehler)

      default: {
        errorMessage.value = 'Unerwarteter Serverfehler. Bitte später erneut versuchen.'
        additionalInformation.value = 'HTTP-Statuscode: ' + response?.status
      }
    }



    if (e instanceof SyntaxError) {
      errorMessage.value = 'Falsche Antwort erhalten. Bitte später erneut versuchen.'
      additionalInformation.value = 'Falscher Content-Type. JSON erwartet, aber ' + contentType.value + ' erhalten. '
    }

/*
  else if (e instanceof TypeError) {
      errorMessage.value = "Server nicht erreichbar. Bitte später erneut versuchen."
      additionalInformation.value = e.toString()
    }

 */

  }
}

init();
</script>

<template>
  <div>
   </div>
    <div v-if="state === State.Error" class="flex justify-center items-center h-screen">
      <div class="errorMessage">
        <p> {{errorMessage}}</p>
        <button class="border border-black p-0.5 text-sm bg-white" @click="toggleInfo()">Mehr Informationen</button>
        <p v-if="showAdditionalInformation">{{ additionalInformation }}</p>
      </div>
    </div>
    <RouterView v-if="state === State.ShowRoutes"/>
</template>



<style>
.errorMessage {
  @apply border border-gray-600 p-4 rounded-md bg-gray-50

}


</style>