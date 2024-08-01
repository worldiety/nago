<script setup lang="ts">
import UiErrorMessage from '@/components/UiErrorMessage.vue';
import {useErrorHandling} from '@/composables/errorhandling';
import type {ComponentInvalidated} from "@/shared/protocol/ora/componentInvalidated";
import {nextTick, onBeforeMount, onMounted, onUnmounted, ref, watch} from "vue";
import type {Component} from "@/shared/protocol/ora/component";
import GenericUi from "@/components/UiGeneric.vue";
import type {NavigationForwardToRequested} from "@/shared/protocol/ora/navigationForwardToRequested";
import type {Event} from '@/shared/protocol/ora/event';
import {useEventBus} from '@/composables/eventBus';
import {useServiceAdapter} from '@/composables/serviceAdapter';
import {EventType} from '@/shared/eventbus/eventType';
import type {ErrorOccurred} from '@/shared/protocol/ora/errorOccurred';
import type {SendMultipleRequested} from "@/shared/protocol/ora/sendMultipleRequested";
import type {Themes} from '@/shared/protocol/ora/themes';
import type {Theme} from '@/shared/protocol/ora/theme';
import {useThemeManager} from '@/shared/themeManager';
import {WindowInfo} from "@/shared/protocol/ora/windowInfo";
import {URI} from "@/shared/protocol/ora/uRI";

enum State {
	Loading,
	ShowUI,
	Error,
}

const eventBus = useEventBus();
const serviceAdapter = useServiceAdapter();
const themeManager = useThemeManager();
const state = ref(State.Loading);
const ui = ref<Component>();

const errorHandler = useErrorHandling();
let configurationPromise: Promise<void> | null = null;

//TODO: Torben baut zukünftig /health ein, der einen 200er und eine json-response zurückgibt, wenn der Service grundsätzlich läuft

async function applyConfiguration(): Promise<void> {
	// establish connection, may be to an existing scope (hold in SPAs memory only to avoid n:1 connection
	// restoration).
	await serviceAdapter.initialize();

	// request and apply configuration
	const config = await serviceAdapter.getConfiguration();
	themeManager.setThemes(config.themes);
	themeManager.applyActiveTheme();
	updateFavicon(config.appIcon)
	sendWindowInfo();
}

function updateFavicon(uri: URI) {
	if (!uri || uri.length == 0) {
		return
	}


	var link = document.querySelector("link[rel~='icon']");
	if (!link) {
		link = document.createElement('link');
		link.rel = 'icon';
		document.head.appendChild(link);
	}

	link.href = uri
}

async function initializeUi(): Promise<void> {
	try {
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
		}, "", null)
		const invalidation = await serviceAdapter.createComponent(factoryId, params)

		eventBus.subscribe(EventType.INVALIDATED, updateUi);
		eventBus.subscribe(EventType.ERROR_OCCURRED, handleError);
		eventBus.subscribe(EventType.NAVIGATE_FORWARD_REQUESTED, navigateForward);
		eventBus.subscribe(EventType.NAVIGATE_BACK_REQUESTED, navigateBack);
		eventBus.subscribe(EventType.NAVIGATE_RELOAD_REQUESTED, navigateReload);
		eventBus.subscribe(EventType.NAVIGATION_RESET_REQUESTED, resetHistory);
		eventBus.subscribe(EventType.SEND_MULTIPLE_REQUESTED, sendMultipleRequested);

		updateUi(invalidation);
	} catch {
		state.value = State.Error;
	}
}

function handleError(event: Event): void {
	//alert((event as ErrorOccurred).message);
	console.log((event as ErrorOccurred).message)
}

function updateUi(event: Event): void {
	if (event.type !== EventType.INVALIDATED) {
		return;
	}
	const componentInvalidated = event as ComponentInvalidated;
	console.log("setting new view tree", componentInvalidated.value)
	ui.value = componentInvalidated.value;
	state.value = State.ShowUI;
}

async function navigateForward(event: Event): Promise<void> {
	console.log("navigate forward", ui.value)
	if (!ui.value) {
		return;
	}

	const navigationForwardToRequested = (event as NavigationForwardToRequested);
	await serviceAdapter.destroyComponent(ui.value?.id)
	const componentInvalidated = await serviceAdapter.createComponent(navigationForwardToRequested.factory, navigationForwardToRequested.values);
	ui.value = componentInvalidated.value;

	let url = `/${navigationForwardToRequested.factory}`
	if (navigationForwardToRequested.values && Object.entries(navigationForwardToRequested.values).length > 0) {
		url += '?';
		Object.entries(navigationForwardToRequested.values).forEach(([key, value], index, array) => {
			url += `${key}=${value}`;
			if (index < array.length - 1) {
				url += '&';
			}
		});
	}
	history.pushState(navigationForwardToRequested, "", url)
}

function navigateBack(): void {
	history.back();
}

function navigateReload(): void {
	location.reload()
}

function resetHistory(event: Event): void {
	// todo this seems not possible in the web
	navigateForward(event)
}

function sendMultipleRequested(evt: Event): void {
	let msg = evt as SendMultipleRequested;
	let res = msg.resources[0];

	let a = document.createElement('a');
	a.href = res.uri;
	a.download = res.name;
	document.body.appendChild(a);
	a.click();
	document.body.removeChild(a);
}

function setTheme(themes: Themes): void {
	let activeTheme: Theme;
	switch (localStorage.getItem('color-theme')) {
		case 'light':
			activeTheme = themes.light;
			break;
		case 'dark':
			activeTheme = themes.dark;
			break;
		default:
			activeTheme = themes.light;
			break;
	}
}

const activeBreakpoint = ref('');

function sendWindowInfo() {
	const breakpoints = {
		sm: 640,
		md: 768,
		lg: 1024,
		xl: 1280,
		'2xl': 1536,
	};

	const lastActiveBreakpoint = activeBreakpoint.value;

	const width = window.innerWidth;
	if (width >= breakpoints['2xl']) activeBreakpoint.value = '2xl';
	else if (width >= breakpoints.xl) activeBreakpoint.value = 'xl';
	else if (width >= breakpoints.lg) activeBreakpoint.value = 'lg';
	else if (width >= breakpoints.md) activeBreakpoint.value = 'md';
	else activeBreakpoint.value = 'sm';

	if (lastActiveBreakpoint == activeBreakpoint.value) {
		// avoid spamming the backend with messages from fluid window resizing
		return
	}

	let currentTheme = localStorage.getItem('color-theme')
	if (!currentTheme) {
		currentTheme = ""
	}

	//console.log("active breakpoint", activeBreakpoint.value)
	const winfo: WindowInfo = {
		width: window.innerWidth,
		height: window.innerHeight,
		density: window.devicePixelRatio,
		sizeClass: activeBreakpoint.value,
		colorScheme: currentTheme,
	}


	serviceAdapter.updateWindowInfo(winfo);
}

function addEventListeners(): void {
	addEventListener("popstate", (event) => {
		if (event.state === null) {
			return
		}

		const req2 = history.state as NavigationForwardToRequested
		if (ui.value) {
			serviceAdapter.destroyComponent(ui.value.id)
		}
		serviceAdapter.createComponent(req2.factory, req2.values).then(invalidation => {
			ui.value = invalidation.value;
		})
	});

	window.addEventListener('resize', function (event) {
		sendWindowInfo();
	});
}

onBeforeMount(() => {
	configurationPromise = applyConfiguration();
});

onMounted(async () => {
	await configurationPromise;
	await initializeUi();
	addEventListeners();

});

onUnmounted(() => {
	serviceAdapter.teardown();
	eventBus.unsubscribe(EventType.INVALIDATED, updateUi);
	eventBus.unsubscribe(EventType.ERROR_OCCURRED, handleError);
	eventBus.unsubscribe(EventType.NAVIGATE_FORWARD_REQUESTED, navigateForward);
	eventBus.unsubscribe(EventType.NAVIGATE_BACK_REQUESTED, navigateBack);
	eventBus.unsubscribe(EventType.NAVIGATION_RESET_REQUESTED, resetHistory);
	eventBus.unsubscribe(EventType.SEND_MULTIPLE_REQUESTED, sendMultipleRequested);
});


//modal dialog support
const anyModalVisible = ref<boolean>(false);
const windowScrollY = ref<number>(0);

// we just watch for changes
// TODO dont know the render timing and states
watch(() => ui.value, (newValue) => {
	if (newValue) {
		if (!anyModalVisible.value) {
			windowScrollY.value = window.scrollY * -1;
			anyModalVisible.value = true;
		}
	} else {
		anyModalVisible.value = false;
		nextTick(() => {
			window.scrollTo(0, windowScrollY.value * -1);
		})
	}
});

</script>

<style scoped>
.modal-container {
	z-index: var(--modal-z-index);
}

.content-container.content-container-freezed {
	@apply fixed left-0 right-0;
	top: var(--content-top-offset);
}
</style>

<template>
	<div v-if="errorHandler.error.value" class="flex h-screen items-center justify-center">
		<UiErrorMessage :error="errorHandler.error.value"></UiErrorMessage>
	</div>

	<div id="ora-modals" class="modal-container fixed inset-0 pointer-events-none" style="--modal-z-index: 40">

	</div>


	<div class="bg-M1 content-container  min-h-screen">
		<!--  <div>Dynamic page information: {{ page }}</div> -->
		<div v-if="state === State.Loading">Loading UI definition…</div>
		<div v-else-if="state === State.Error">Failed to fetch UI definition.</div>
		<generic-ui v-else-if="state === State.ShowUI && ui" :ui="ui"/>
		<div v-else>Empty UI</div>
	</div>

</template>
