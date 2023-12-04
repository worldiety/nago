<!--
    This page will build its UI dynamically according to the PageConfiguration loaded from the server.
-->
<script lang="ts" setup>
import {useRoute, useRouter} from 'vue-router';
import type {PageConfiguration} from '@/shared/model';
import {provide, ref, watch} from 'vue';
import GenericUi from '@/components/UiGeneric.vue';
import {useHttp} from '@/shared/http';
import {Invalidation, LiveComponent, LiveMessage} from "@/shared/livemsg";


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


// Provide the current UiDescription to all child elements.
// https://vuejs.org/guide/components/provide-inject.html
provide('ui', ui);
provide('ws', ws);

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
        state.value = State.ShowUI;
        invalidationResp.value = msg
        return
      case "HistoryPushState":
        history.pushState({}, "", msg.pageId + "?" + encodeQueryData(msg.state))
        location.reload()
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
          <!--
            Modal panel, show/hide based on modal state.

            Entering: "ease-out duration-300"
              From: "opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
              To: "opacity-100 translate-y-0 sm:scale-100"
            Leaving: "ease-in duration-200"
              From: "opacity-100 translate-y-0 sm:scale-100"
              To: "opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
          -->
          <div class="relative transform overflow-hidden sm:my-8 sm:w-full sm:max-w-lg rounded-lg">


              <generic-ui :ui="modal" :ws="ws!"/>

            <!--
            <div class="bg-white px-4 pb-4 pt-5 sm:p-6 sm:pb-4">
              <div class="sm:flex sm:items-start">
                <div class="mx-auto flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full bg-red-100 sm:mx-0 sm:h-10 sm:w-10">
                  <svg class="h-6 w-6 text-red-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
                  </svg>
                </div>
                <div class="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left">
                  <h3 class="text-base font-semibold leading-6 text-gray-900" id="modal-title">Deactivate account</h3>
                  <div class="mt-2">
                    <p class="text-sm text-gray-500">Are you sure you want to deactivate your account? All of your data will be permanently removed. This action cannot be undone.</p>
                  </div>
                </div>
              </div>
            </div>
            <div class="bg-gray-50 px-4 py-3 sm:flex sm:flex-row-reverse sm:px-6">
              <button type="button" class="inline-flex w-full justify-center rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-500 sm:ml-3 sm:w-auto">Deactivate</button>
              <button type="button" class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto">Cancel</button>
            </div>
            -->


          </div>
        </div>
      </div>
    </div>

  </div>

  <div>


    <!--  <div>Dynamic page information: {{ page }}</div> -->
    <div v-if="state === State.Loading">Loading UI definition…</div>
    <div v-else-if="state === State.Error">Failed to fetch UI definition.</div>
    <generic-ui v-else-if="state === State.ShowUI && ui" :ui="ui" :ws="ws!"/>
    <div v-else>Empty UI</div>
  </div>


</template>
