<script setup lang="ts">
import UiErrorMessage from '@/components/UiErrorMessage.vue';
import {useErrorHandling} from '@/composables/errorhandling';
import {ComponentInvalidated} from "@/shared/protocol/gen/componentInvalidated";
import {ErrorOccurred} from "@/shared/protocol/gen/errorOccurred";
import {onUnmounted, provide, ref} from "vue";
import {useNetworkStore} from "@/stores/networkStore";
import {Component} from "@/shared/protocol/gen/component";
import GenericUi from "@/components/UiGeneric.vue";
import {NavigationForwardToRequested} from "@/shared/protocol/gen/navigationForwardToRequested";
import type { Event } from '@/shared/protocol/gen/event';

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

const errorHandler = useErrorHandling();

//TODO: Torben baut zukünftig /health ein, der einen 200er und eine json-response zurückgibt, wenn der Service grundsätzlich läuft

async function init(): Promise<void> {
	try {
		// establish connection, may be to an existing scope (hold in SPAs memory only to avoid n:1 connection
		// restoration).
		await networkStore.initialize();

		// configure the scope with color scheme and locale
		// TODO: connect this to the scheme and locale picker for accessibility
		let cfg = await networkStore.getConfiguration("light", navigator.languages[0])
		console.log("my config", cfg)

		// create a new component (which is likely a page but not necessarily)
		let factoryId = window.location.pathname.substring(1);
		if (factoryId.length === 0) {
			factoryId = "." // this is by ora definition the root page
		}
		console.log(`factory: ${factoryId}`)
		let params: Record<string, string> = {};
		new URLSearchParams(window.location.search).forEach((value, key) => {
			params[key] = value
		})
		history.replaceState({
			factory:factoryId,
			values:params,
		},"",null)
		let invalidation = await networkStore.newComponent(factoryId, params)
		console.log("my render tree", invalidation)

		// todo is this the right place? when to remove the subscriber?
		networkStore.addUnrequestedEventSubscriber((event: Event) => {
			switch (event.type) {
				case "ComponentInvalidated":
					ui.value = (event as ComponentInvalidated).value
					break
				case "ErrorOccurred":
					alert((event as ErrorOccurred).message)
					break
				case "NavigationForwardToRequested":
					const req = (event as NavigationForwardToRequested);
					networkStore.destroyComponent(ui.value?.id)
					networkStore.newComponent(req.factory, req.values).then(invalidation => {
						ui.value = invalidation.value;
					})

					let url = `/${req.factory}?`
					Object.entries(req.values).forEach(([key, value]) => {
						url += `${key}=${value}&`
					});
					history.pushState(req, "", url)
					break
				case "NavigationBackRequested":
					history.back()
					break
				default:
					console.log("ignored unhandled event", event)
			}
		})

		ui.value = invalidation.value;
		state.value = State.ShowUI;
		console.log("app init done")
	} catch {
		state.value = State.Error;
	}
}

init();
addEventListener("popstate",(event)=>{
	if (event.state===null){
		console.log("bogus history")
		return
	}

	let req2 = history.state as NavigationForwardToRequested
	networkStore.destroyComponent(ui.value?.id)
	networkStore.newComponent(req2.factory, req2.values).then(invalidation => {
		ui.value = invalidation.value;
	})
})

onUnmounted(() => {
	networkStore.teardown();
});

</script>

<template>
	<div v-if="errorHandler.error.value" class="flex h-screen items-center justify-center">
		<UiErrorMessage :error="errorHandler.error.value"></UiErrorMessage>
	</div>


	<div>
		<!--  <div>Dynamic page information: {{ page }}</div> -->
		<div v-if="state === State.Loading">Loading UI definition…</div>
		<div v-else-if="state === State.Error">Failed to fetch UI definition.</div>
		<generic-ui v-else-if="state === State.ShowUI && ui" :ui="ui"/>
		<div v-else>Empty UI</div>
	</div>

</template>
