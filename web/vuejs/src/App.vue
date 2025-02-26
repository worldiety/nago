<script setup lang="ts">
import {nextTick, onBeforeMount, onMounted, onUnmounted, ref, watch} from 'vue';
import {useUploadRepository} from '@/api/upload/uploadRepository';
import GenericUi from '@/components/UiGeneric.vue';
import ConnectingChannelOverlay from '@/components/overlays/ConnectingChannelOverlay.vue';
import ConnectionLostOverlay from '@/components/overlays/ConnectionLostOverlay.vue';
import {useServiceAdapter} from '@/composables/serviceAdapter';
import {
	applyRootViewState,
	clipboardWriteText,
	getWindowInfo,
	lastRID,
	navigateForward,
	nextRID,
	onScopeConfigurationChanged,
	openHttpFlow,
	openHttpLink,
	requestRootViewAllocation,
	requestRootViewRendering,
	requestScopeConfigurationChange,
	setTheme,
	triggerFileDownload,
	triggerFileUpload,
	windowInfoChanged,
} from '@/eventhandling';
import ConnectionHandler from '@/shared/network/connectionHandler';
import {ConnectionState} from '@/shared/network/connectionState';
import {
	ClipboardWriteTextRequested,
	Component,
	ErrorRootViewAllocationRequired,
	FileImportRequested,
	NavigationBackRequested,
	NavigationForwardToRequested,
	NavigationReloadRequested,
	NavigationResetRequested,
	OpenHttpFlow,
	OpenHttpLink,
	RootViewInvalidated,
	RootViewRenderingRequested,
	ScopeConfigurationChanged,
	SendMultipleRequested,
	ThemeRequested,
} from '@/shared/proto/nprotoc_gen';
import {useThemeManager} from '@/shared/themeManager';

enum State {
	Loading,
	ShowUI,
	Error,
}

const serviceAdapter = useServiceAdapter();
const themeManager = useThemeManager();
const state = ref(State.Loading);
const ui = ref<Component>();
const componentKey = ref(0);

const connected = ref(true);

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
	addConnectionListeners();

	requestScopeConfigurationChange(serviceAdapter, themeManager);
	fixHistoryInit();

	ConnectionHandler.addEventListener((evt) => {
		//console.log('app received nago event', evt);
		if (evt instanceof ScopeConfigurationChanged) {
			onScopeConfigurationChanged(themeManager, evt);
			return;
		}

		if (evt instanceof RootViewInvalidated) {
			if (evt.rID.value != 0 && evt.rID.value < lastRID().value) {
				console.log(
					'received outdated root view rendering, discarding',
					'expected',
					lastRID().value,
					'received',
					evt.rID.value
				);
				return;
			}

			ui.value = evt.root;
			state.value = State.ShowUI;
			return;
		}

		if (evt instanceof ErrorRootViewAllocationRequired) {
			requestRootViewAllocation(serviceAdapter, themeManager.activeLocale);
			return;
		}

		if (evt instanceof SendMultipleRequested) {
			triggerFileDownload(evt);
			return;
		}

		if (evt instanceof FileImportRequested) {
			triggerFileUpload(uploadRepository, evt);
			return;
		}

		if (evt instanceof NavigationForwardToRequested) {
			navigateForward(serviceAdapter, evt);
			return;
		}

		if (evt instanceof OpenHttpLink) {
			openHttpLink(evt);
			return;
		}

		if (evt instanceof OpenHttpFlow) {
			openHttpFlow(evt);
			return;
		}

		if (evt instanceof ThemeRequested) {
			setTheme(serviceAdapter, themeManager, evt);
			return;
		}

		if (evt instanceof NavigationBackRequested) {
			history.back();
			return;
		}

		if (evt instanceof NavigationReloadRequested) {
			location.reload();
			return;
		}

		if (evt instanceof NavigationResetRequested) {
			// todo this seems not possible in the web
			navigateForward(serviceAdapter, new NavigationForwardToRequested(evt.rootView, evt.values));
			return;
		}

		if (evt instanceof ClipboardWriteTextRequested) {
			clipboardWriteText(evt);
			return;
		}

		console.log('unhandled event from backend', evt);
	});

	requestRootViewRendering(serviceAdapter);
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
			location.reload();
		} else {
			console.log('restore cookie: unexpected result', response);
		}
	});
}

function fixHistoryInit() {
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
}

const uploadRepository = useUploadRepository();

const activeBreakpoint = ref(-1);

function addEventListeners(): void {
	addEventListener('popstate', (event) => {
		if (event.state === null) {
			return;
		}

		console.log('pop state', event);
		applyRootViewState(serviceAdapter, history.state);
	});

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
	console.log('connection changed', connected.value);
	if (connected.value) {
		console.log('websocket connected, poke server');
		// always send the window info changed, otherwise if the server lost its state, the rendering
		// has the wrong dimensions and breakspoints apply wrong
		windowInfoChanged(serviceAdapter, themeManager);
		serviceAdapter.sendEvent(new RootViewRenderingRequested(nextRID()));
	}
}

function addConnectionListeners(): void {
	ConnectionHandler.addConnectionChangeListener(onConnectionChange);
}

onBeforeMount(() => {
	configurationPromise = applyConfiguration();
});

onMounted(async () => {
});

onUnmounted(() => {
	serviceAdapter.teardown();
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
	<ConnectionLostOverlay v-if="!connected"/>
	<ConnectingChannelOverlay v-if="state === State.Loading"/>

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
		<generic-ui v-else-if="state === State.ShowUI && ui" :ui="ui"/>
		<div v-else>Empty UI</div>
	</div>
</template>
