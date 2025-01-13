<script setup lang="ts">
import {nextTick, onBeforeMount, onMounted, onUnmounted, ref, watch} from 'vue';
import {useUploadRepository} from '@/api/upload/uploadRepository';
import UiErrorMessage from '@/components/UiErrorMessage.vue';
import GenericUi from '@/components/UiGeneric.vue';
import ConnectionLostOverlay from '@/components/overlays/ConnectionLostOverlay.vue';
import {useErrorHandling} from '@/composables/errorhandling';
import {useEventBus} from '@/composables/eventBus';
import {useServiceAdapter} from '@/composables/serviceAdapter';
import {EventType} from '@/shared/eventbus/eventType';
import ConnectionHandler from '@/shared/network/connectionHandler';
import {ConnectionState} from '@/shared/network/connectionState';
import type {Component} from '@/shared/protocol/ora/component';
import type {ComponentInvalidated} from '@/shared/protocol/ora/componentInvalidated';
import type {ErrorOccurred} from '@/shared/protocol/ora/errorOccurred';
import type {Event} from '@/shared/protocol/ora/event';
import {FileImportRequested} from '@/shared/protocol/ora/fileImportRequested';
import type {NavigationForwardToRequested} from '@/shared/protocol/ora/navigationForwardToRequested';
import {OpenRequested} from '@/shared/protocol/ora/openRequested';
import type {SendMultipleRequested} from '@/shared/protocol/ora/sendMultipleRequested';
import type {Theme} from '@/shared/protocol/ora/theme';
import {ThemeRequested} from '@/shared/protocol/ora/themeRequested';
import type {Themes} from '@/shared/protocol/ora/themes';
import {URI} from '@/shared/protocol/ora/uRI';
import {WindowInfo} from '@/shared/protocol/ora/windowInfo';
import {useThemeManager} from '@/shared/themeManager';

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
const componentKey = ref(0);

const connected = ref(true);

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
	updateFavicon(config.appIcon);
	sendWindowInfo(false);
}

function updateFavicon(uri: URI) {
	if (!uri || uri.length == 0) {
		return;
	}

	var link = document.querySelector("link[rel~='icon']");
	if (!link) {
		link = document.createElement('link');
		link.rel = 'icon';
		document.head.appendChild(link);
	}

	link.href = uri;
}

async function initializeUi(): Promise<void> {
	try {
		// create a new component (which is likely a page but not necessarily)
		let factoryId = window.location.pathname.substring(1);
		if (factoryId.length === 0) {
			factoryId = '.'; // this is by ora definition the root page
		}
		const params: Record<string, string> = {};
		new URLSearchParams(window.location.search).forEach((value, key) => {
			params[key] = value;
		});
		history.replaceState(
			{
				factory: factoryId,
				values: params,
			},
			'',
			null
		);
		const invalidation = await serviceAdapter.createComponent(factoryId, params);

		eventBus.subscribe(EventType.INVALIDATED, updateUi);
		eventBus.subscribe(EventType.ERROR_OCCURRED, handleError);
		eventBus.subscribe(EventType.NAVIGATE_FORWARD_REQUESTED, navigateForward);
		eventBus.subscribe(EventType.NAVIGATE_BACK_REQUESTED, navigateBack);
		eventBus.subscribe(EventType.NAVIGATE_RELOAD_REQUESTED, navigateReload);
		eventBus.subscribe(EventType.NAVIGATION_RESET_REQUESTED, resetHistory);
		eventBus.subscribe(EventType.SEND_MULTIPLE_REQUESTED, sendMultipleRequested);
		eventBus.subscribe(EventType.FILE_IMPORT_REQUESTED, fileImportRequested);
		eventBus.subscribe(EventType.WindowInfoChanged, sendWindowInfo);
		eventBus.subscribe(EventType.THEME_REQUESTED, themeRequested);
		eventBus.subscribe(EventType.OPEN_REQUESTED, openRequested);
		eventBus.subscribe(EventType.ServerStateLost, serverStateLost);

		updateUi(invalidation);
	} catch {
		state.value = State.Error;
	}
}

function serverStateLost(): void{
	// the most important point is, that the server got a new version
	// thus, the correct reaction is to reload everything, because this frontend may have changed.
	navigateReload();
}

function handleError(event: Event): void {
	//alert((event as ErrorOccurred).message);
	console.log((event as ErrorOccurred).message);
}

function updateUi(event: Event): void {
	if (event.type !== EventType.INVALIDATED) {
		return;
	}
	const componentInvalidated = event as ComponentInvalidated;
	console.log('setting new view tree', componentInvalidated.value);
	ui.value = componentInvalidated.value;
	state.value = State.ShowUI;
}

async function navigateForward(event: Event): Promise<void> {
	console.log('navigate forward', ui.value);
	if (!ui.value) {
		return;
	}

	const navigationForwardToRequested = event as NavigationForwardToRequested;
	await serviceAdapter.destroyComponent(ui.value?.id);
	const componentInvalidated = await serviceAdapter.createComponent(
		navigationForwardToRequested.factory,
		navigationForwardToRequested.values
	);
	ui.value = componentInvalidated.value;

	componentKey.value += 1;
	console.log('componentkey', componentKey.value);

	let url = `/${navigationForwardToRequested.factory}`;
	if (navigationForwardToRequested.values && Object.entries(navigationForwardToRequested.values).length > 0) {
		url += '?';
		Object.entries(navigationForwardToRequested.values).forEach(([key, value], index, array) => {
			url += `${key}=${value}`;
			if (index < array.length - 1) {
				url += '&';
			}
		});
	}
	history.pushState(navigationForwardToRequested, '', url);
}

function navigateBack(): void {
	history.back();
}

function navigateReload(): void {
	location.reload();
}

function resetHistory(event: Event): void {
	// todo this seems not possible in the web
	navigateForward(event);
}

const uploadRepository = useUploadRepository();

function openRequested(evt: Event): void {
	let msg = evt as OpenRequested;
	if (!msg.options) {
		open(msg.resource);
		return;
	}

	switch (msg.options['_type']) {
		case 'http-flow':
			let redirectTarget = msg.options['redirectTarget'];
			let redirectNavigation = msg.options['redirectNavigation'];

			console.log('http-flow', redirectTarget, redirectNavigation);

			window.location.href = msg.resource;
			break;
		case 'http-link':
			let target = msg.options['target'];
			window.open(msg.resource, target);
			break;
		default:
			open(msg.resource);
	}

	sendWindowInfo(true);
}

function themeRequested(evt: Event): void {
	let msg = evt as ThemeRequested;

	switch (msg.theme) {
		case 'light':
			themeManager.applyLightmodeTheme();
			break;
		case 'dark':
			themeManager.applyDarkmodeTheme();
			break;
		default:
			console.log('unknown theme', msg.theme);
	}

	sendWindowInfo(true);
}

function fileImportRequested(evt: Event): void {
	let msg = evt as FileImportRequested;
	let input = document.createElement('input');
	input.className = 'hidden';
	input.type = 'file';
	input.id = msg.id;
	input.multiple = msg.multiple;
	input.onchange = (event) => {
		const item = event.target as HTMLInputElement;
		if (!item.files) {
			return;
		}
		for (let i = 0; i < item.files.length; i++) {
			uploadRepository.fetchUpload(
				item.files[i],
				msg.id,
				0,
				msg.scopeID,
				(uploauploadId: string, progress: number, total: number) => {
					console.log('progress', progress);
				},
				(uploadId) => {},
				(uploadId) => {},
				(uploadId) => {
					console.log('upload failed');
				}
			);
		}
	};
	if (msg.allowedMimeTypes) {
		input.accept = msg.allowedMimeTypes.join(',');
	}
	document.body.appendChild(input);
	input.showPicker();
	//	input.click()
	document.body.removeChild(input);
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

function sendWindowInfo(force: boolean = true) {
	const breakpoints = {
		'sm': 640,
		'md': 768,
		'lg': 1024,
		'xl': 1280,
		'2xl': 1536,
	};

	const lastActiveBreakpoint = activeBreakpoint.value;

	const width = window.innerWidth;
	if (width >= breakpoints['2xl']) activeBreakpoint.value = '2xl';
	else if (width >= breakpoints.xl) activeBreakpoint.value = 'xl';
	else if (width >= breakpoints.lg) activeBreakpoint.value = 'lg';
	else if (width >= breakpoints.md) activeBreakpoint.value = 'md';
	else activeBreakpoint.value = 'sm';

	if (!force && lastActiveBreakpoint == activeBreakpoint.value) {
		// avoid spamming the backend with messages from fluid window resizing
		return;
	}

	let currentTheme = localStorage.getItem('color-theme');
	if (!currentTheme) {
		currentTheme = '';
	}

	//console.log("active breakpoint", activeBreakpoint.value)
	const winfo: WindowInfo = {
		width: window.innerWidth,
		height: window.innerHeight,
		density: window.devicePixelRatio,
		sizeClass: activeBreakpoint.value,
		colorScheme: currentTheme,
	};

	serviceAdapter.updateWindowInfo(winfo);
}

function addEventListeners(): void {
	addEventListener('popstate', (event) => {
		if (event.state === null) {
			return;
		}

		const req2 = history.state as NavigationForwardToRequested;
		if (ui.value) {
			serviceAdapter.destroyComponent(ui.value.id);
		}
		serviceAdapter.createComponent(req2.factory, req2.values).then((invalidation) => {
			ui.value = invalidation.value;
		});
	});

	window.addEventListener('resize', function (event) {
		sendWindowInfo(false);
	});
}

function onConnectionChange(connectionState: ConnectionState): void {
	connected.value = connectionState.connected;
	if (connected.value){
		// trigger a re-render, TODO introduce something like an invalidate event
		serviceAdapter.executeFunctions(-1)
	}
}

function addConnectionListeners(): void {
	ConnectionHandler.addConnectionChangeListener(onConnectionChange);
}

onBeforeMount(() => {
	configurationPromise = applyConfiguration();
});

onMounted(async () => {
	await configurationPromise;
	await initializeUi();
	addEventListeners();
	addConnectionListeners();
});

onUnmounted(() => {
	serviceAdapter.teardown();
	eventBus.unsubscribe(EventType.INVALIDATED, updateUi);
	eventBus.unsubscribe(EventType.ERROR_OCCURRED, handleError);
	eventBus.unsubscribe(EventType.NAVIGATE_FORWARD_REQUESTED, navigateForward);
	eventBus.unsubscribe(EventType.NAVIGATE_BACK_REQUESTED, navigateBack);
	eventBus.unsubscribe(EventType.NAVIGATION_RESET_REQUESTED, resetHistory);
	eventBus.unsubscribe(EventType.SEND_MULTIPLE_REQUESTED, sendMultipleRequested);
	eventBus.unsubscribe(EventType.FILE_IMPORT_REQUESTED, fileImportRequested);
	eventBus.unsubscribe(EventType.THEME_REQUESTED, themeRequested);
	eventBus.unsubscribe(EventType.OPEN_REQUESTED, openRequested);
});

//modal dialog support
const anyModalVisible = ref<boolean>(false);
const windowScrollY = ref<number>(0);

// we just watch for changes
// TODO dont know the render timing and states
watch(
	() => ui.value,
	(newValue) => {
		if (newValue) {
			if (!anyModalVisible.value) {
				windowScrollY.value = window.scrollY * -1;
				anyModalVisible.value = true;
			}
		} else {
			anyModalVisible.value = false;
			nextTick(() => {
				window.scrollTo(0, windowScrollY.value * -1);
			});
		}
	}
);
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
	<ConnectionLostOverlay v-if="!connected" />

	<div v-if="errorHandler.error.value" class="flex h-screen items-center justify-center">
		<UiErrorMessage :error="errorHandler.error.value"></UiErrorMessage>
	</div>

	<div
		id="ora-overlay"
		class="modal-container fixed inset-0 pointer-events-none z-40"
		style="--modal-z-index: 35"
	></div>

	<div id="ora-modals" class="modal-container fixed inset-0 pointer-events-none" style="--modal-z-index: 40"></div>

	<div class="bg-M1 content-container min-h-screen">
		<!--  <div>Dynamic page information: {{ page }}</div> -->
		<div v-if="state === State.Loading">Loading UI definition…</div>
		<div v-else-if="state === State.Error">Failed to fetch UI definition.</div>
		<generic-ui v-else-if="state === State.ShowUI && ui" :ui="ui" />
		<div v-else>Empty UI</div>
	</div>
</template>
