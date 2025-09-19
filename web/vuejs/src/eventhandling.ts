/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { UploadRepository } from '@/api/upload/uploadRepository';
import { Channel } from '@/shared/network/serviceAdapter';
import {
	CallMediaDevicesEnumerate,
	CallRequested,
	CallResolved,
	ClipboardWriteTextRequested,
	ColorSchemeValues,
	FileImportRequested,
	Fonts,
	Locale,
	MediaDevice,
	MediaDeviceKindValues,
	MediaDevices,
	NavigationForwardToRequested,
	OpenHttpFlow,
	OpenHttpLink,
	RID,
	RetMediaDevicesEnumerate,
	RetMediaDevicesPermissionsError,
	RootViewAllocationRequested,
	RootViewID,
	RootViewParameters,
	RootViewRenderingRequested,
	ScopeConfigurationChangeRequested,
	ScopeConfigurationChanged,
	SendMultipleRequested,
	ThemeRequested,
	URI,
	WindowInfo,
	WindowInfoChanged,
	WindowSizeClass,
	WindowSizeClassValues,
} from '@/shared/proto/nprotoc_gen';
import ThemeManager, { ThemeKey } from '@/shared/themeManager';

let nextRequestTracingID: number = 1;

/**
 * nextRID increments the global request tracing number and returns it.
 * This is not functionally relevant, but it may help for debugging event order related questions.
 */
export function nextRID(): RID {
	nextRequestTracingID++;
	return nextRequestTracingID;
}

// lastRID returns that last returned request/response tracing number.
export function lastRID(): RID {
	return nextRequestTracingID;
}

/**
 * windowInfoChanged emits the according event into the channel. There is logic behind it to avoid
 * sending redundant or spamming changed events.
 */
export function windowInfoChanged(chan: Channel, themeManager: ThemeManager) {
	const windowInfo = getWindowInfo(themeManager);
	chan.sendEvent(new WindowInfoChanged(windowInfo, nextRID()));
}

/**
 * getWindowInfo calculates the current WindowInfo and returns it.
 */
export function getWindowInfo(themeManager: ThemeManager): WindowInfo {
	let windowInfo = new WindowInfo();
	windowInfo.density = window.devicePixelRatio;
	windowInfo.width = window.innerWidth;
	windowInfo.height = window.innerHeight;
	windowInfo.sizeClass = currentSizeClass();
	windowInfo.userAgent = navigator.userAgent;

	if (themeManager.getActiveThemeKey() === ThemeKey.DARK) {
		windowInfo.colorScheme = ColorSchemeValues.Dark;
	} else {
		windowInfo.colorScheme = ColorSchemeValues.Light;
	}

	return windowInfo;
}

/**
 * requestRootViewRendering emits the according request blindly to the backend. This may either result
 * in various error variants or the actual rendering.
 * Note that, depending on the way how a scope is re-connected, there may be still a view allocated which just waits
 * to be displayed. We can never know that.
 */
export function requestRootViewRendering(chan: Channel) {
	chan.sendEvent(new RootViewRenderingRequested());
}

/**
 * requestRootViewAllocation emits the according event based on the current window location.
 * It also replaces the current history state, so that going back and forth will result in correct navigation behavior.
 * The backend will trigger a rendering by specification automatically.
 */
export function requestRootViewAllocation(chan: Channel, locale: Locale) {
	let rootViewID = requiredRootViewID();
	let rootViewParams = requiredRootViewParameter();

	history.replaceState(new NavigationForwardToRequested(rootViewID, rootViewParams), '', null);

	chan.sendEvent(new RootViewAllocationRequested(locale, rootViewID, nextRID(), rootViewParams));
}

/**
 * requestConfigurationChange sends an initiative event to the backend. Usually, this should only happen
 * once after initialization. Note, that there is a special event just for [WindowInfoChanged].
 */
export function requestScopeConfigurationChange(chan: Channel, themeManager: ThemeManager) {
	let evt = new ScopeConfigurationChangeRequested();
	evt.windowInfo = getWindowInfo(themeManager);

	evt.acceptLanguage = getLocale();
	chan.sendEvent(evt);
}

/**
 * getLocale returns whatever the browser thinks, the locale/language the user wants.
 */
export function getLocale(): Locale {
	if (navigator.languages && navigator.languages.length > 0) {
		return navigator.languages[0] as Locale;
	}

	return navigator.language as Locale;
}

/**
 * onScopeConfigurationChanged is called if the backend has changed its configuration. This will at least happen
 * after the backend has processed a [ScopeConfigurationChangeRequested] event.
 */
export function onScopeConfigurationChanged(themeManager: ThemeManager, evt: ScopeConfigurationChanged) {
	if (!evt.themes) {
		return;
	}
	themeManager.setThemes(evt.themes);
	themeManager.applyActiveTheme();
	if (evt.activeLocale) {
		themeManager.activeLocale = evt.activeLocale;
	}

	updateFavicon(evt.appIcon);

	if (evt.fonts) {
		loadFonts(evt.fonts);
	}

	console.log('onScopeConfigurationChanged', evt);
}

const alreadyLoadedFonts = new Map<string, boolean>();

/**
 * loadFonts inspects whatever is in the given fonts and loads only those faces, which are unknown.
 */
function loadFonts(fonts: Fonts) {
	if (fonts.defaultFontFace) {
		document.documentElement.style.setProperty('font-family', `'${fonts.defaultFontFace}', sans-serif`);
	}

	if (!fonts.faces) {
		return;
	}

	fonts.faces.value.forEach((faceDef) => {
		if (alreadyLoadedFonts.has(faceDef.source!)) {
			return;
		}

		const fontFace = new FontFace(faceDef.family!, `url(${faceDef.source!})`, {
			weight: faceDef.weight ? faceDef.weight : '400',
			style: faceDef.style ? faceDef.style : 'normal',
		});

		fontFace
			.load()
			.then((value) => {
				document.fonts.add(fontFace);
				console.log(`extra font ${JSON.stringify(faceDef)} loaded`);
			})
			.catch((err) => {
				const debug = JSON.stringify(faceDef);
				console.log(`failed to load font ${debug}:`, err);
			});

		alreadyLoadedFonts.set(faceDef.source!, true);
	});
}

/**
 * currentSizeClass determines (in a partially hardcoded way) which tailwind break point matches the Nago size class.
 * This is error-prone, because we cannot read out the tailwind config here (as far as I know), so if the tailwind
 * break points change, this must be updated by hand to be consistent.
 */
function currentSizeClass(): WindowSizeClass {
	const breakpoints = {
		'sm': 640,
		'md': 768,
		'lg': 1024,
		'xl': 1280,
		'2xl': 1536,
	};

	let wsc: WindowSizeClass;
	const width = window.innerWidth;

	if (width >= breakpoints['2xl']) wsc = WindowSizeClassValues.SizeClass2XL;
	else if (width >= breakpoints.xl) wsc = WindowSizeClassValues.SizeClassXL;
	else if (width >= breakpoints.lg) wsc = WindowSizeClassValues.SizeClassLarge;
	else if (width >= breakpoints.md) wsc = WindowSizeClassValues.SizeClassMedium;
	else wsc = WindowSizeClassValues.SizeClassSmall;

	return wsc;
}

/**
 * updateFavicon installs the given uri (if not empty) into the document, replacing any other favicon.
 */
function updateFavicon(uri?: URI) {
	if (!uri) {
		return;
	}

	let link = document.querySelector("link[rel~='icon']") as HTMLLinkElement;
	if (!link) {
		link = document.createElement('link');
		link.rel = 'icon';
		document.head.appendChild(link);
	}

	link.href = uri;
}

/**
 * requiredRootViewID returns the current root view based on the current window location path.
 * If pathname is empty, the Nago defined index identifier "." is returned.
 */
function requiredRootViewID(): RootViewID {
	let factoryId = window.location.pathname.substring(1);
	if (factoryId.length === 0) {
		factoryId = '.'; // this is by ora definition the root page
	}

	return factoryId;
}

/**
 * requiredRootViewParameter return the current expected root view parameters based on the current window location
 * query parameters. This is how Nago root view parameters are defined to work in the web. These params must
 * be stateless and safe for bookmarking and must not expose secrets. But that is the responsibility of the backend.
 */
function requiredRootViewParameter(): RootViewParameters {
	let params = new RootViewParameters();
	new URLSearchParams(window.location.search).forEach((value, key) => {
		params.value.set(key, value);
	});

	return params;
}

/**
 * triggerFileDownload applies some hacks to simulate a user-requested file download by inserting fake
 * nodes and clicking on them. After some tries, this seems to be the most stable behavior across all browsers.
 * @param evt
 */
export function triggerFileDownload(evt: SendMultipleRequested): void {
	if (!evt.resources) {
		return;
	}

	let res = evt.resources.value[0];
	let a = document.createElement('a');
	a.href = res.uRI!;
	a.download = res.name!;
	document.body.appendChild(a);
	a.click();
	document.body.removeChild(a);
}

/**
 * triggerFileUpload applies some hack to simulate a user-requested file upload by inserting fake
 * nodes and showing the native browser file picker.
 * @param uploadRepository
 * @param evt
 */
export async function triggerFileUpload(uploadRepository: UploadRepository, evt: FileImportRequested): Promise<void> {
	let input = document.createElement('input');
	input.className = 'hidden';
	input.type = 'file';
	input.id = evt.iD!;
	input.multiple = evt.multiple!;
	input.onchange = async (event) => {
		const item = event.target as HTMLInputElement;
		if (!item.files) {
			return;
		}
		for (let i = 0; i < item.files.length; i++) {
			// design decision: disallow parallel uploads which complicates backend onCompletionHandler design for developers.
			// Without serializing, the naive backend implementation will receive file handlers concurrently and in "no-time"
			// but block a "decade" which results in a lot of (at least) logical data races caused by appending to
			// slices.
			await uploadRepository.fetchUpload(
				item.files[i],
				evt.iD!,
				0,
				evt.scopeID!,
				(uploauploadId: string, progress: number, total: number) => {
					console.log('progress', progress);
				},
				(uploadId) => {
					// upload finished
				},
				(uploadId) => {
					// upload aborted
				},
				(uploadId) => {
					console.log('upload failed');
				}
			);
		}
	};
	if (evt.allowedMimeTypes?.value) {
		input.accept = evt.allowedMimeTypes.value.join(',');
	}
	document.body.appendChild(input);
	input.showPicker();
	//	input.click() // this does not work properly on safari
	document.body.removeChild(input);
}

/**
 * navigateForward issues a RootViewAllocationRequested to the backend and updates the browser history stack.
 * @param chan
 * @param evt
 */
export function navigateForward(chan: Channel, evt: NavigationForwardToRequested): void {
	//console.log('!!!!', evt);
	let url = `/${evt.rootView!}`;
	if (evt.values) {
		url += '?';
		let idx = 0;
		evt.values.value.forEach((value, key) => {
			if (!evt.values?.value.size) {
				return;
			}
			url += `${key}=${value}`;
			if (idx < evt.values.value.size - 1) {
				url += '&';
			}
			idx++;
		});
	}

	if (evt.target === '_blank') {
		// special case without manipulating our history, instead open in new tab
		window.open(url, '_blank');
		return;
	}

	// otherwise handle locally
	chan.sendEvent(new RootViewAllocationRequested(getLocale(), evt.rootView, nextRID(), evt.values));
	if (lastScrolledRootView !== evt.rootView) {
		lastScrolledRootView = evt.rootView;
		nextInvalidationScrollsTopFlag = true;
	}

	history.pushState(evt, '', url);
}

// nextInvalidationScrollsTop is set by navigation events and tells if a redraw must trigger a scroll to top,
// e.g. because it is really a new page.
var nextInvalidationScrollsTopFlag: boolean;
var lastScrolledRootView: RootViewID | undefined;

export function nextInvalidationScrollsTop(): boolean {
	let tmp = nextInvalidationScrollsTopFlag;
	nextInvalidationScrollsTopFlag = false;
	console.log('reset next scroll flag');
	return tmp;
}

// scrollToTop issues a scrolling to the top of the window. Note that posting may cause a flickering.
export function scrollToTop(post: boolean) {
	console.log('scroll to top');
	if (post) {
		setTimeout(() => window.scrollTo(0, 0), 0);
	} else {
		window.scrollTo(0, 0);
	}
}

/**
 * applyRootViewState applies a new root view based on the given state.
 * @param chan
 * @param state
 */
export function applyRootViewState(chan: Channel, state: any) {
	const evt = state as NavigationForwardToRequested;
	let req = new RootViewAllocationRequested();
	// important: evt/history.state may be in broken state, due to the way how javascript deserializes the state
	// it is NOT of NavigationForwardToRequested anymore
	console.log('applyRootViewState from history', state);
	if (evt.rootView) {
		req.factory = evt.rootView;
	}

	if (req.factory === '') {
		req.factory = '.';
	}

	if (evt.values && evt.values.value) {
		req.values = new RootViewParameters(evt.values.value);
	}

	req.locale = getLocale();
	req.rID = nextRID();

	lastScrolledRootView = evt.rootView;
	nextInvalidationScrollsTopFlag = true;

	chan.sendEvent(req);
}

/**
 * openHttpLink trivially calls the browsers window open function.
 * @param evt
 */
export function openHttpLink(evt: OpenHttpLink) {
	window.open(evt.url, evt.target);
}

/**
 * openHttpFlow replaces the current location and saves any http flow session to peek through
 * the CSRF protection for later redirects back.
 * @param evt
 */
export function openHttpFlow(evt: OpenHttpFlow) {
	localStorage.setItem('http-flow-session', evt.session!);
	window.location.href = evt.url!;
}

/**
 * setTheme updates the theme from the event and triggers the theme manager.
 * @param chan
 * @param themeManager
 * @param evt
 */
export function setTheme(chan: Channel, themeManager: ThemeManager, evt: ThemeRequested): void {
	switch (evt.theme) {
		case 'light':
			themeManager.applyLightmodeTheme();
			break;
		case 'dark':
			themeManager.applyDarkmodeTheme();
			break;
		default:
			console.log('unknown theme', evt.theme);
	}

	windowInfoChanged(chan, themeManager);
}

export function clipboardWriteText(evt: ClipboardWriteTextRequested) {
	return navigator.clipboard
		.writeText(evt.text!)
		.then(() => {
			console.log('text written to clipboard');
			return '';
		})
		.catch((reason) => {
			console.log('failed to copy text into clipboard', reason);
			if (reason.toString().lastIndexOf('NotAllowed') >= 0) {
				return evt.text!;
			}
			return '';
		});
}

export async function callRequested(chan: Channel, evt: CallRequested) {
	if (evt.call instanceof CallMediaDevicesEnumerate) {
		await callMediaDevicesEnumerate(chan, evt, evt.call);
	}
}

async function callMediaDevicesEnumerate(chan: Channel, evt: CallRequested, args: CallMediaDevicesEnumerate) {
	const withAudio = args.withAudio ?? false;
	const withVideo = args.withVideo ?? false;
	try {
		await navigator.mediaDevices.getUserMedia({
			audio: withAudio,
			video: withVideo,
		});
		console.log('media device get user media success', 'video', withVideo, 'audio', withAudio);
	} catch (e) {
		console.warn("Couldn't get requested permissions", e);
		chan.sendEvent(new CallResolved(evt.callPtr, new RetMediaDevicesPermissionsError(e.toString(), 403)));
		return;
	}

	const devices = await navigator.mediaDevices.enumerateDevices();
	const tmp: MediaDevice[] = devices
		.map(mapMediaDeviceInfoToMediaDevice)
		.filter((d): d is MediaDevice => filterByMediaKind(d, withVideo, withAudio));

	console.log('got media devices enumeration', tmp);

	chan.sendEvent(new CallResolved(evt.callPtr, new RetMediaDevicesEnumerate(new MediaDevices(tmp)), nextRID()));
}

function filterByMediaKind(device: MediaDevice | undefined, withVideo: boolean, withAudio: boolean): boolean {
	if (!device) return false;

	switch (device.kind) {
		case MediaDeviceKindValues.AudioInput:
		case MediaDeviceKindValues.AudioOutput:
			return withAudio;
		case MediaDeviceKindValues.VideoInput:
			return withVideo;
		default:
			return false;
	}
}

function mapMediaDeviceInfoToMediaDevice(d: MediaDeviceInfo): MediaDevice | undefined {
	if (!d.deviceId || d.deviceId.length === 0) {
		// do not map unidentified devices
		return undefined;
	}

	return new MediaDevice(d.deviceId, d.groupId, d.label, getMediaDeviceKindFromMediaDeviceInfo(d));
}

function getMediaDeviceKindFromMediaDeviceInfo(device: MediaDeviceInfo): MediaDeviceKindValues {
	switch (device.kind) {
		case 'audioinput':
			return MediaDeviceKindValues.AudioInput;
		case 'audiooutput':
			return MediaDeviceKindValues.AudioOutput;
		case 'videoinput':
			return MediaDeviceKindValues.VideoInput;
	}
}
