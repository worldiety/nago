<!--
    This page will build its UI dynamically according to the PageConfiguration loaded from the server.
-->
<script lang="ts" setup>
import {onUnmounted, provide, ref, watch} from 'vue';
import GenericUi from '@/components/UiGeneric.vue';
import {useNetworkStore} from '@/stores/networkStore';
import {ComponentInvalidated} from "@/shared/protocol/gen/componentInvalidated";
import {Component} from "@/shared/protocol/gen/component";
import {ErrorOccurred} from "@/shared/protocol/gen/errorOccurred";

enum State {
	Loading,
	ShowUI,
	Error,
}

const networkStore = useNetworkStore();


const state = ref(State.Loading);
const ui = ref<Component>();

// Provide the current UiDescription to all child elements.
// https://vuejs.org/guide/components/provide-inject.html
provide('ui', ui);

async function init() {
	try {
		const pageUrl = import.meta.env.VITE_HOST_BACKEND + 'api/v1/ui/page' + document.location.pathname;


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
					break
				case "ErrorOccurred":
					alert((evt as ErrorOccurred).message)
					break
			}
		})

		ui.value = invalidation.value;
		state.value = State.ShowUI;
		console.log("old page async init done",ui)
	} catch {
		state.value = State.Error;
	}
}

init();

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

		<div v-if="state === State.ShowUI && ui" v-for="modal in ui.modals.v">
			<div class="fixed inset-0 z-50 bg-gray-700 bg-opacity-75 transition-opacity"></div>

			<div class="fixed inset-0 z-50 w-screen overflow-y-auto">
				<div class="flex min-h-full items-center justify-center p-4 text-center sm:p-0">
					<div class="relative transform overflow-hidden rounded-lg sm:my-8 sm:w-full sm:max-w-lg">
						<generic-ui   :ui="modal"/>
					</div>
				</div>
			</div>
		</div>
	</div>
	<div>
		<!--  <div>Dynamic page information: {{ page }}</div> -->
		<div v-if="state === State.Loading">Loading UI definition…</div>
		<div v-else-if="state === State.Error">Failed to fetch UI definition.</div>
		<generic-ui v-else-if="state === State.ShowUI && ui" :ui="ui" />
		<div v-else>Empty UI</div>
	</div>
</template>
