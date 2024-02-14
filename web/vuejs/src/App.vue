

<script setup lang="ts">
import {RouterView, useRoute, useRouter} from 'vue-router';
import Page from '@/views/Page.vue';
import {ref} from 'vue';
import type {PagesConfiguration} from '@/shared/model';
import {useAuth} from "@/stores/auth";
import {UserManager} from "oidc-client-ts";
import {LiveMessage} from "@/shared/livemsg";
import axios from 'axios';

const router = useRouter();
const route = useRoute();

enum State {
  LoadingRoutes,
  ShowRoutes,
  Error,
}

const auth = useAuth();
const state = ref(State.LoadingRoutes);

const errorMessage = ref()
const additionalInformation = ref()

const isOnline = navigator.onLine
const captivePortal = ref()
const captiveOk = ref()
const showAdditionalInformation = ref(false)
const contentType = ref()

const checkCaptivePortal = async () => {
  try {
    captivePortal.value = await axios.get('https://captive.apple.com')
    return captiveOk.value = true

  } catch (error) {
    state.value = State.Error
    return captiveOk.value = false

  }
}

const checkInternetConnection = () => {
  if (!isOnline) {
    errorMessage.value = "Keine Internetverbindung vorhanden. Bitte Verbindung überprüfen.";
    additionalInformation.value = "Router und Kabel überprüfen. Eventuell WLAN-Verbindung wiederherstellen.";
    return false;
  }
  return true;
};


const toggleInfo = () => {
  showAdditionalInformation.value = !showAdditionalInformation.value
}

async function init() {
  if (!checkInternetConnection()) {
    return
  }

  captiveOk.value = await checkCaptivePortal()

  if (captiveOk.value === false) {
    errorMessage.value = "Keine Verbindung möglich. Bitte Rechte prüfen."
    additionalInformation.value = "Captive Portal Check fehlgeschlagen."
  }


  try {
    const response = await fetch(import.meta.env.VITE_HOST_BACKEND + 'api/v1/ui/application');
    contentType.value = response.headers.get('Content-Type');

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
    await router.replace(route);

    state.value = State.ShowRoutes;


    if (router.currentRoute.value.path==="/" && app.index != null && app.index != "") {
      console.log("app requires index rewrite to ", app.index)
      router.replace(app.index)
    }



  } catch (e) {
    state.value = State.Error


    if (e instanceof SyntaxError) {
      errorMessage.value = 'Falsche Antwort erhalten. Bitte später erneut versuchen.'
      additionalInformation.value = 'Falscher Content-Type. JSON erwartet, aber ' + contentType.value + ' erhalten.'
    } else if (e instanceof TypeError) {
      errorMessage.value = "Server nicht erreichbar. Bitte später erneut versuchen."
      additionalInformation.value = e
    }
  }
}



init();
</script>

<template>
  <div>
   </div>
    <div class="flex justify-center items-center h-screen" v-if="state === State.Error || !isOnline || !captiveOk">
      <div class="errorMessage">
        <p> {{errorMessage}}</p>
        <button class="border border-black p-0.5 text-sm bg-white" @click="toggleInfo()">Mehr Informationen</button>
        <p v-if="showAdditionalInformation">{{ additionalInformation }}</p>
      </div>
    </div>
    <RouterView v-if="state === State.ShowRoutes && isOnline && captiveOk"/>
</template>



<style>
.errorMessage {
  @apply border border-gray-600 p-4 rounded-md bg-gray-50

}


</style>