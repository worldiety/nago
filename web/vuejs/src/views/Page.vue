<!--
    This page will build its UI dynamically according to the PageConfiguration loaded from the server.
-->
<script lang="ts" setup>
import {useRoute, useRouter} from 'vue-router';
import type {PageConfiguration} from '@/shared/model';
import {onUnmounted, provide, ref, watch} from 'vue';
import GenericUi from '@/components/UiGeneric.vue';
import {useHttp} from '@/shared/http';
import {
  CallBatch,
  CallServerFunc,
  ClientHello,
  Invalidation,
  LiveComponent,
  LiveMessage,
  LivePage,
  UpdateJWT
} from "@/shared/livemsg";
import {useAuth, UserChangedCallbacks} from "@/stores/auth";


enum State {
  Loading,
  ShowUI,
  Error,
}

const route = useRoute();
const router = useRouter();

const page = route.meta.page as PageConfiguration;

const http = useHttp();

const state = ref(State.Loading);
const ui = ref<LiveComponent>();
const invalidationResp = ref<Invalidation>({});
const ws = ref<WebSocket>();
const livePage = ref<LivePage>({})


// Provide the current UiDescription to all child elements.
// https://vuejs.org/guide/components/provide-inject.html
provide('ui', ui);
provide('ws', ws);
provide('livePage', livePage);

async function init() {
  try {

    // const router = useRouter()
    const pageUrl = import.meta.env.VITE_HOST_BACKEND + "api/v1/ui/page" + router.currentRoute.value.path//page.link.slice(1);
    console.log("i'm in init", pageUrl)
    /* const response = await http.request(pageUrl);
     ui.value = await response.json();
     state.value = State.ShowUI;
     console.log(pageUrl);
     console.log('got value', ui.value);*/
    connectWebSocket()
  } catch {
    state.value = State.Error;
  }
}


init();

watch(route, () => {
  state.value = State.Loading
  init()

})

onUnmounted(()=>{
  ws.value?.close()
})

function retry() {
  setTimeout(connectWebSocket, 2000)
}

function connectWebSocket() {
  console.log("trying ws open")

  let myPort = import.meta.env.VITE_WS_BACKEND_PORT
  if (myPort === "") {
    myPort = window.location.port
  }



  let wsurl = "ws://" + window.location.hostname + ":" + myPort + "/wire?_pid=" + window.location.pathname.substring(1)
  let queryString = window.location.search.substring(1)
  wsurl += "&" + queryString

  console.log("open websocket ->" + wsurl)
  let lws = new WebSocket(wsurl);

  lws.onopen = function (evt) {
    console.log("OPEN");
    ws.value = lws
    sendHello()
   // sendUser()

  }
  lws.onclose = function (evt) {
    console.log("CLOSE");
    retry()
  }
  lws.onmessage = function (evt) {
    console.log("RESPONSE: " + evt.data);

    let msg: LiveMessage = JSON.parse(evt.data)
    console.log(msg)


    switch (msg.type) {
      case "Invalidation":
        ui.value = msg.root
        livePage.value = msg
        state.value = State.ShowUI;
        invalidationResp.value = msg
        return
      case "HistoryPushState":
        history.pushState({}, "", msg.pageId + "?" + encodeQueryData(msg.state))
        location.reload() // TODO this does not always work like the refresh button, because the websocket and everything is not reconnected
        console.log("push state")
        return
      case "HistoryBack":
        history.back();
        return
    }


  }
  lws.onerror = function (evt) {
    console.log("ERROR: " + evt);
    state.value = State.Error;
    retry()
  }


  console.log("ws ???")

  UserChangedCallbacks.push(user => sendUser())
}

function sendUser(){
  const auth = useAuth();

  const updateJWT: UpdateJWT = {
    type: "updateJWT",
    token: `${auth.user?.access_token}`,
    OIDCName:"Keycloak",
  }

  const callTx: CallBatch = {
    tx: [updateJWT]
  }

  ws.value?.send(JSON.stringify(callTx))
}

function sendHello(){
  const auth = useAuth();

  const hello: ClientHello = {
    type: "hello",
    auth: {
      keycloak: `${auth.user?.access_token}`,
    },
  }

  const callTx: CallBatch = {
    tx: [hello]
  }

  console.log(JSON.stringify(callTx))

  ws.value?.send(JSON.stringify(callTx))
}

function encodeQueryData(data) {
  const ret = [];
  for (let d in data)
    ret.push(encodeURIComponent(d) + '=' + encodeURIComponent(data[d]));
  return ret.join('&');
}

</script>

<template>

  <div class="relative z-50" aria-labelledby="modal-title" role="dialog" aria-modal="true">
    <!--
      Background backdrop, show/hide based on modal state.

      Entering: "ease-out duration-300"
        From: "opacity-0"
        To: "opacity-100"
      Leaving: "ease-in duration-200"
        From: "opacity-100"
        To: "opacity-0"
    -->

    <div v-for="modal in invalidationResp.modals">
      <div class="fixed z-50 inset-0 bg-gray-700 bg-opacity-75 transition-opacity"></div>

      <div class="fixed inset-0 z-50 w-screen overflow-y-auto">
        <div class="flex min-h-full  justify-center p-4 text-center items-center sm:p-0">

          <div class="relative transform overflow-hidden sm:my-8 sm:w-full sm:max-w-lg rounded-lg">


              <generic-ui :ui="modal" :ws="ws!" :page="livePage"/>




          </div>
        </div>
      </div>
    </div>

  </div>

  <div>




    <!--  <div>Dynamic page information: {{ page }}</div> -->
    <div v-if="state === State.Loading">Loading UI definition…</div>
    <div v-else-if="state === State.Error">Failed to fetch UI definition.</div>
    <generic-ui v-else-if="state === State.ShowUI && ui" :ui="ui" :ws="ws!" :page="livePage"/>
    <div v-else>Empty UI</div>
  </div>


</template>
