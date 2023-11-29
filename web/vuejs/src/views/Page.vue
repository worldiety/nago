<!--
    This page will build its UI dynamically according to the PageConfiguration loaded from the server.
-->
<script lang="ts" setup>
import {useRoute, useRouter} from 'vue-router';
import type {PageConfiguration, Scaffold} from '@/shared/model';
import {onMounted, onUpdated, provide, ref, watch} from 'vue';
import GenericUi from '@/components/UiGeneric.vue';
import {useHttp} from '@/shared/http';
import {LiveMessage} from "@/shared/livemsg";


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
const ui = ref<Scaffold>();
const ws = ref<WebSocket>();

// Provide the current UiDescription to all child elements.
// https://vuejs.org/guide/components/provide-inject.html
provide('ui', ui);
provide('ws',ws);

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

function retry(){
  setTimeout(connectWebSocket,2000)
}

function connectWebSocket() {
  console.log("trying ws open")

  let lws = new WebSocket("ws://localhost:3000/wire");

  lws.onopen = function (evt) {
    console.log("OPEN");
    ws.value=lws
  }
  lws.onclose = function (evt) {
    console.log("CLOSE");
    retry()
  }
  lws.onmessage = function (evt) {
    console.log("RESPONSE: " + evt.data);

    let msg: LiveMessage = JSON.parse(evt.data)
    console.log(msg)

    ui.value = msg.root
    state.value = State.ShowUI;

  }
  lws.onerror = function (evt) {
    console.log("ERROR: " + evt);
    state.value = State.Error;
    retry()
  }


  console.log("ws ???")


}

function initDarkModeToggle(){
  var themeToggleDarkIcon = document.getElementById('theme-toggle-dark-icon');
  var themeToggleLightIcon = document.getElementById('theme-toggle-light-icon');

  // Change the icons inside the button based on previous settings
  if (localStorage.getItem('color-theme') === 'dark' || (!('color-theme' in localStorage) && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    themeToggleLightIcon.classList.remove('hidden');
  } else {
    themeToggleDarkIcon.classList.remove('hidden');
  }

  var themeToggleBtn = document.getElementById('theme-toggle');

  themeToggleBtn.addEventListener('click', function() {

    // toggle icons inside button
    themeToggleDarkIcon.classList.toggle('hidden');
    themeToggleLightIcon.classList.toggle('hidden');

    // if set via local storage previously
    if (localStorage.getItem('color-theme')) {
      if (localStorage.getItem('color-theme') === 'light') {
        document.documentElement.classList.add('dark');
        localStorage.setItem('color-theme', 'dark');
      } else {
        document.documentElement.classList.remove('dark');
        localStorage.setItem('color-theme', 'light');
      }

      // if NOT set via local storage previously
    } else {
      if (document.documentElement.classList.contains('dark')) {
        document.documentElement.classList.remove('dark');
        localStorage.setItem('color-theme', 'light');
      } else {
        document.documentElement.classList.add('dark');
        localStorage.setItem('color-theme', 'dark');
      }
    }

  });
}

onMounted(() => {
 // initDarkModeToggle() //TODO
})


</script>

<template>


  <div >


    <!--  <div>Dynamic page information: {{ page }}</div> -->
    <div v-if="state === State.Loading">Loading UI definition…</div>
    <div v-else-if="state === State.Error">Failed to fetch UI definition.</div>
    <generic-ui v-else-if="state === State.ShowUI && ui" :ui="ui" :ws="ws!"/>
    <div v-else>Empty UI</div>
  </div>
</template>
