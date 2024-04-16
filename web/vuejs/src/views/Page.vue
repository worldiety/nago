<!--
    This page will build its UI dynamically according to the PageConfiguration loaded from the server.
-->
<script lang="ts" setup>
import {useRoute, useRouter} from 'vue-router';
import {onUnmounted, provide, ref, watch} from 'vue';
import GenericUi from '@/components/UiGeneric.vue';
import {useNetworkStore} from '@/stores/networkStore';
import type {PageConfiguration} from '@/shared/model/pageConfiguration';
import type {LiveComponent} from '@/shared/model/liveComponent';
import type {Invalidation} from '@/shared/model/invalidation';
import type {LivePage} from '@/shared/model/livePage';
import type {LiveMessage} from '@/shared/model/liveMessage';
import {ComponentInvalidated} from "@/shared/protocol/gen/componentInvalidated";
import {Component} from "@/shared/protocol/gen/component";

enum State {
	Loading,
	ShowUI,
	Error,
}

const route = useRoute();
const router = useRouter();
const networkStore = useNetworkStore();

const page = route.meta.page as PageConfiguration;

const state = ref(State.Loading);
const ui = ref<Component>();
const invalidationResp = ref<Invalidation>({});
const ws = ref<WebSocket>();
const livePage = ref<LivePage>({});

// Provide the current UiDescription to all child elements.
// https://vuejs.org/guide/components/provide-inject.html
provide('ui', ui);
provide('ws', ws);
provide('livePage', livePage);

async function init() {
	try {
		// const router = useRouter()
		const pageUrl = import.meta.env.VITE_HOST_BACKEND + 'api/v1/ui/page' + router.currentRoute.value.path; //page.link.slice(1);


		// establish connection, may be to an existing scope (hold in SPAs memory only to avoid n:1 connection
		// restoration).
		await networkStore.initialize();
		console.log("network initialized", pageUrl);

		// configure the scope with color scheme and locale
		// TODO: connect this to the scheme and locale picker for accessibility
		let cfg = await networkStore.getConfiguration("light", navigator.languages[0])
		console.log("my config", cfg)

		// create a new component (which is likely a page but not necessarily)
		let factoryId = window.location.pathname.substring(1);
		let params = new Map<string, string>();
		new URLSearchParams(window.location.search).forEach((value, key) => {
			params.set(key, value)
		})
		let invalidation = await networkStore.newComponent(factoryId, params)
		console.log("my render tree", invalidation)

		// todo is this the right place? when to remove the subscriber?
		networkStore.addUnprocessedEventSubscriber(evt => {
			switch (evt.type){
				case "ComponentInvalidated":
					ui.value=(evt as ComponentInvalidated).value
			}
		})

		ui.value = invalidation.value;
		livePage.value = invalidation.value;
		invalidationResp.value = invalidation.value;
		state.value = State.ShowUI;
		console.log("old page async init done",ui)
	} catch {
		state.value = State.Error;
	}
}

init();

watch(route, () => {
	state.value = State.Loading;
	init();
});

onUnmounted(() => {
	networkStore.teardown();
});

console.log("old page")
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

		<div v-if="state === State.ShowUI && ui" v-for="modal in invalidationResp.modals.v">
			<div class="fixed inset-0 z-50 bg-gray-700 bg-opacity-75 transition-opacity"></div>

			<div class="fixed inset-0 z-50 w-screen overflow-y-auto">
				<div class="flex min-h-full items-center justify-center p-4 text-center sm:p-0">
					<div class="relative transform overflow-hidden rounded-lg sm:my-8 sm:w-full sm:max-w-lg">
						<generic-ui   :ui="modal" :page="livePage"/>
					</div>
				</div>
			</div>
		</div>
	</div>
	<div>
		<!--  <div>Dynamic page information: {{ page }}</div> -->
		<div v-if="state === State.Loading">Loading UI definition…</div>
		<div v-else-if="state === State.Error">Failed to fetch UI definition.</div>
		<generic-ui v-else-if="state === State.ShowUI && ui" :ui="ui" :page="livePage"/>
		<div v-else>Empty UI</div>
	</div>
</template>
