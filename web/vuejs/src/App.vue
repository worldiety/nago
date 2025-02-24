<script setup lang="ts">
import { nextTick, onBeforeMount, onMounted, onUnmounted, ref, watch } from 'vue';
import { useUploadRepository } from '@/api/upload/uploadRepository';
import UiErrorMessage from '@/components/UiErrorMessage.vue';
import GenericUi from '@/components/UiGeneric.vue';
import ConnectingChannelOverlay from '@/components/overlays/ConnectingChannelOverlay.vue';
import ConnectionLostOverlay from '@/components/overlays/ConnectionLostOverlay.vue';
import { useErrorHandling } from '@/composables/errorhandling';
import { useEventBus } from '@/composables/eventBus';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import {
	getWindowInfo,
	onScopeConfigurationChanged,
	requestRootViewAllocation,
	requestRootViewRendering,
	requestScopeConfigurationChange,
	windowInfoChanged,
} from '@/eventhandling';
import { EventType } from '@/shared/eventbus/eventType';
import ConnectionHandler from '@/shared/network/connectionHandler';
import { ConnectionState } from '@/shared/network/connectionState';
import {
	Component,
	ErrorRootViewAllocationRequired,
	RootViewInvalidated,
	ScopeConfigurationChanged,
	WindowInfo,
} from '@/shared/proto/nprotoc_gen';
import type { ComponentInvalidated } from '@/shared/protocol/ora/componentInvalidated';
import type { ErrorOccurred } from '@/shared/protocol/ora/errorOccurred';
import type { Event } from '@/shared/protocol/ora/event';
import { FileImportRequested } from '@/shared/protocol/ora/fileImportRequested';
import type { NavigationForwardToRequested } from '@/shared/protocol/ora/navigationForwardToRequested';
import { OpenRequested } from '@/shared/protocol/ora/openRequested';
import type { SendMultipleRequested } from '@/shared/protocol/ora/sendMultipleRequested';
import type { Theme } from '@/shared/protocol/ora/theme';
import { ThemeRequested } from '@/shared/protocol/ora/themeRequested';
import type { Themes } from '@/shared/protocol/ora/themes';
import { useThemeManager } from '@/shared/themeManager';

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
	// this is part of the (oauth2) security process, which removes our actual session due to strict cookie rules.
	// thus we saved the end-to-end encrypted cookie in our local storage and ask the server to restore it.
	let httpFlowSession = localStorage.getItem('http-flow-session');
	if (httpFlowSession) {
		await restoreCookie(httpFlowSession);
		return;
	}

	// establish connection, may be to an existing scope (hold in SPAs memory only to avoid n:1 connection
	// restoration).

	await serviceAdapter.initialize();
	addEventListeners();

	requestScopeConfigurationChange(serviceAdapter, themeManager);

	ConnectionHandler.addEventListener((evt) => {
		console.log('app received nago event', evt);
		if (evt instanceof ScopeConfigurationChanged) {
			onScopeConfigurationChanged(themeManager, evt);
			return;
		}

		if (evt instanceof RootViewInvalidated) {
			ui.value = evt.root;
			state.value = State.ShowUI;
			return;
		}

		if (evt instanceof ErrorRootViewAllocationRequired) {
			requestRootViewAllocation(serviceAdapter, themeManager.activeLocale);
		}
	});

	requestRootViewRendering(serviceAdapter);
	/*
		// request and apply configuration
		const config = await serviceAdapter.getConfiguration();
		themeManager.setThemes(config.themes);
		themeManager.applyActiveTheme();
		updateFavicon(config.appIcon);
		sendWindowInfo(false);*/
}

function restoreCookie(sessionID: string) {
	return fetch('/api/nago/v1/session/restore', {
		method: 'POST',
		headers: {
			'Content-Type': 'text/plain',
		},
		body: sessionID,
	}).then((response) => {
		if (response.ok) {
			localStorage.removeItem('http-flow-session');
			console.log('completed cookie restoration process');
			navigateReload();
		} else {
			console.log('restore cookie: unexpected result', response);
		}
	});
}

async function initializeUi(): Promise<void> {
	try {
		// these must be registered before requested, especially the navigation things.
		eventBus.subscribe(EventType.NAVIGATE_FORWARD_REQUESTED, navigateForward);
		eventBus.subscribe(EventType.NAVIGATE_BACK_REQUESTED, navigateBack);
		eventBus.subscribe(EventType.NAVIGATE_RELOAD_REQUESTED, navigateReload);
		eventBus.subscribe(EventType.NAVIGATION_RESET_REQUESTED, resetHistory);

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

function serverStateLost(): void {
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

	const navigationForwardToRequested = event as NavigationForwardToRequested;
	if (ui.value) {
		await serviceAdapter.destroyComponent(ui.value?.id);
	}
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
			let session = msg.options['session'];
			localStorage.setItem('http-flow-session', session);

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

	windowInfoChanged(serviceAdapter);
	//sendWindowInfo(true);
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

	windowInfoChanged(serviceAdapter);
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

const activeBreakpoint = ref(-1);

function addEventListeners(): void {
	/*addEventListener('popstate', (event) => {
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
	});*/

	window.addEventListener('resize', function (event) {
		const info = getWindowInfo(themeManager);
		if (info.sizeClass.value === activeBreakpoint.value) {
			// avoid spamming the backend with messages from fluid window resizing
			return;
		}

		activeBreakpoint.value = info.sizeClass.value;
		windowInfoChanged(serviceAdapter, themeManager);
	});
}

function onConnectionChange(connectionState: ConnectionState): void {
	connected.value = connectionState.connected;
	if (connected.value) {
		// trigger a re-render, TODO use ComponentInvalidatedRequested
		console.log('TODO websocket connected, poke server');
		//serviceAdapter.executeFunctions(-1)
	}
}

function addConnectionListeners(): void {
	ConnectionHandler.addConnectionChangeListener(onConnectionChange);
}

onBeforeMount(() => {
	configurationPromise = applyConfiguration();
});

onMounted(async () => {
	//await configurationPromise;
	//await initializeUi();
	//addEventListeners();
	//addConnectionListeners();
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
	<ConnectingChannelOverlay v-if="state === State.Loading" />

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
		<div v-if="state === State.Loading">Warte auf Websocket-Verbindung...</div>
		<div v-else-if="state === State.Error">Failed to fetch UI definition.</div>
		<generic-ui v-else-if="state === State.ShowUI && ui" :ui="ui" />
		<div v-else>Empty UI</div>
	</div>
</template>
