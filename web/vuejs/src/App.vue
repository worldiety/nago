<script setup lang="ts">
import {RouterView, useRoute, useRouter} from 'vue-router';
import Page from '@/views/Page.vue';
import {ref} from 'vue';
import {useAuth} from "@/stores/authStore";
import {UserManager} from "oidc-client-ts";
import type { PagesConfiguration } from '@/shared/model/pagesConfiguration';

enum State {
  LoadingRoutes,
  ShowRoutes,
  Error,
}

const router = useRouter();
const route = useRoute();
const auth = useAuth();
const state = ref(State.LoadingRoutes);

async function init() {
  try {

    const response = await fetch(import.meta.env.VITE_HOST_BACKEND + 'api/v1/ui/application');
    const app: PagesConfiguration = await response.json();

    if (app.oidc?.length>0){
      /*auth.init(new UserManager({
        authority: 'http://localhost:8080/realms/master',
        client_id: 'testclientid',
        redirect_uri: 'http://localhost:8090/oauth',
        post_logout_redirect_uri: 'http://localhost:8090',
      }))*/
      const provider = app.oidc.at(0);
      if (provider) {
        auth.init(new UserManager({
          authority: provider.authority,
          client_id: provider.clientID,
          redirect_uri: provider.redirectURL,
          post_logout_redirect_uri: provider.postLogoutRedirectUri,
        }));
      }
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
    console.log(e);
    state.value = State.Error;
  }
}

init();
</script>

<template>
  <div>
    <div v-if="state === State.LoadingRoutes">Loading…</div>
    <div v-if="state === State.Error">Routes could not be loaded.</div>
    <RouterView v-if="state === State.ShowRoutes"/>
  </div>
</template>
