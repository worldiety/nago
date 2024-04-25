<script setup lang="ts">
import UiErrorMessage from '@/components/UiErrorMessage.vue';
import { useErrorHandling } from '@/composables/errorhandling';
import type { ComponentInvalidated } from "@/shared/protocol/gen/componentInvalidated";
import { onUnmounted, ref } from "vue";
import type { Component } from "@/shared/protocol/gen/component";
import GenericUi from "@/components/UiGeneric.vue";
import type { NavigationForwardToRequested } from "@/shared/protocol/gen/navigationForwardToRequested";
import type { Event } from '@/shared/protocol/gen/event';
import { useEventBus } from '@/composables/eventBus';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { EventType } from '@/shared/eventbus/eventType';
import { ErrorOccurred } from '@/shared/protocol/gen/errorOccurred';

enum State {
	Loading,
	ShowUI,
	Error,
}

const eventBus = useEventBus();
const serviceAdapter = useServiceAdapter();
const state = ref(State.Loading);
const ui = ref<Component>();

const errorHandler = useErrorHandling();

//TODO: Torben baut zukünftig /health ein, der einen 200er und eine json-response zurückgibt, wenn der Service grundsätzlich läuft

async function init(): Promise<void> {
	try {
		// establish connection, may be to an existing scope (hold in SPAs memory only to avoid n:1 connection
		// restoration).
		await serviceAdapter.initialize();

		// create a new component (which is likely a page but not necessarily)
		let factoryId = window.location.pathname.substring(1);
		if (factoryId.length === 0) {
			factoryId = "." // this is by ora definition the root page
		}
		const params: Record<string, string> = {};
		new URLSearchParams(window.location.search).forEach((value, key) => {
			params[key] = value
		})
		history.replaceState({
			factory: factoryId,
			values: params,
		},"",null)
		const invalidation = await serviceAdapter.createComponent(factoryId, params)

		eventBus.subscribe(EventType.INVALIDATED, updateUi);
		eventBus.subscribe(EventType.ERROR_OCCURRED, handleError);
		eventBus.subscribe(EventType.NAVIGATE_FORWARD_REQUESTED, navigateForward);
		eventBus.subscribe(EventType.NAVIGATE_BACK_REQUESTED, navigateBack);

		updateUi(invalidation);
	} catch {
		state.value = State.Error;
	}
}

function handleError(event: Event): void {
	alert((event as ErrorOccurred).message);
}

function updateUi(event: Event): void {
	if (event.type !== EventType.INVALIDATED) {
		return;
	}
	const componentInvalidated = event as ComponentInvalidated;
	ui.value = componentInvalidated.value;
	state.value = State.ShowUI;
}

async function navigateForward(event: Event): void {
	if (!ui.value) {
		return;
	}
	const req = (event as NavigationForwardToRequested);
	await serviceAdapter.destroyComponent(ui.value?.id)
	const componentInvalidated = await serviceAdapter.createComponent(req.factory, req.values);
	ui.value = componentInvalidated.value;

	let url = `/${req.factory}?`
	Object.entries(req.values).forEach(([key, value]) => {
		url += `${key}=${value}&`
	});
	history.pushState(req, "", url)
}

function navigateBack(): void {
	history.back();
}

init();
addEventListener("popstate",(event)=>{
	if (event.state===null){
		return
	}

	const req2 = history.state as NavigationForwardToRequested
	if (ui.value) {
		serviceAdapter.destroyComponent(ui.value.id)
	}
	serviceAdapter.createComponent(req2.factory, req2.values).then(invalidation => {
		ui.value = invalidation.value;
	})
})

onUnmounted(() => {
	serviceAdapter.teardown();
	eventBus.unsubscribe(EventType.INVALIDATED, updateUi);
	eventBus.unsubscribe(EventType.ERROR_OCCURRED, handleError);
	eventBus.unsubscribe(EventType.NAVIGATE_FORWARD_REQUESTED, navigateForward);
	eventBus.unsubscribe(EventType.NAVIGATE_BACK_REQUESTED, navigateBack);
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
