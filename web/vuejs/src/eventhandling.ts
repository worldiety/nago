import {Channel} from '@/shared/network/serviceAdapter';
import {
	ColorScheme,
	ColorSchemeValues,
	Density,
	DP,
	FileImportRequested,
	Locale,
	NavigationForwardToRequested,
	RID,
	RootViewAllocationRequested,
	RootViewID,
	RootViewParameters,
	RootViewRenderingRequested,
	ScopeConfigurationChanged,
	ScopeConfigurationChangeRequested,
	SendMultipleRequested,
	Str,
	WindowInfo,
	WindowInfoChanged,
	WindowSizeClass,
	WindowSizeClassValues,
} from '@/shared/proto/nprotoc_gen';
import {URI} from '@/shared/protocol/ora/uRI';
import ThemeManager, {ThemeKey} from '@/shared/themeManager';
import {UploadRepository} from "@/api/upload/uploadRepository";

let nextRequestTracingID: number = 1;

/**
 * nextRID increments the global request tracing number and returns it.
 * This is not functionally relevant, but it may help for debugging event order related questions.
 */
export function nextRID(): RID {
	nextRequestTracingID++;
	return new RID(nextRequestTracingID);
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
	windowInfo.density = new Density(window.devicePixelRatio);
	windowInfo.width = new DP(window.innerWidth);
	windowInfo.height = new DP(window.innerHeight);
	windowInfo.sizeClass = currentSizeClass();

	if (themeManager.getActiveThemeKey() === ThemeKey.DARK) {
		windowInfo.colorScheme = new ColorScheme(ColorSchemeValues.Dark);
	} else {
		windowInfo.colorScheme = new ColorScheme(ColorSchemeValues.Light);
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

	history.replaceState(
		{
			factory: rootViewID,
			values: rootViewParams,
		},
		'',
		null
	);

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
	return new Locale(navigator.language || navigator.languages[0]);
}

/**
 * onScopeConfigurationChanged is called if the backend has changed its configuration. This will at least happen
 * after the backend has processed a [ScopeConfigurationChangeRequested] event.
 */
export function onScopeConfigurationChanged(themeManager: ThemeManager, evt: ScopeConfigurationChanged) {
	themeManager.setThemes(evt.themes);
	themeManager.applyActiveTheme();
	themeManager.activeLocale = evt.activeLocale;
	updateFavicon(evt.appIcon.toString());
	console.log('onScopeConfigurationChanged', evt);
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

	if (width >= breakpoints['2xl']) wsc = new WindowSizeClass(WindowSizeClassValues.SizeClass2XL);
	else if (width >= breakpoints.xl) wsc = new WindowSizeClass(WindowSizeClassValues.SizeClassXL);
	else if (width >= breakpoints.lg) wsc = new WindowSizeClass(WindowSizeClassValues.SizeClassLarge);
	else if (width >= breakpoints.md) wsc = new WindowSizeClass(WindowSizeClassValues.SizeClassMedium);
	else wsc = new WindowSizeClass(WindowSizeClassValues.SizeClassSmall);

	return wsc;
}

/**
 * updateFavicon installs the given uri (if not empty) into the document, replacing any other favicon.
 */
function updateFavicon(uri: URI) {
	if (!uri || uri.length == 0) {
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

	return new RootViewID(factoryId);
}

/**
 * requiredRootViewParameter return the current expected root view parameters based on the current window location
 * query parameters. This is how Nago root view parameters are defined to work in the web. These params must
 * be stateless and safe for bookmarking and must not expose secrets. But that is the responsibility of the backend.
 */
function requiredRootViewParameter(): RootViewParameters {
	let params = new RootViewParameters();
	new URLSearchParams(window.location.search).forEach((value, key) => {
		params.value.set(new Str(key), new Str(value));
	});

	return params;
}


/**
 * triggerFileDownload applies some hacks to simulate a user-requested file download by inserting fake
 * nodes and clicking on them. After some tries, this seems to be the most stable behavior across all browsers.
 * @param evt
 */
export function triggerFileDownload(evt: SendMultipleRequested): void {
	let res = evt.resources.value[0];
	let a = document.createElement('a');
	a.href = res.uRI.value;
	a.download = res.name.value;
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
export function triggerFileUpload(uploadRepository: UploadRepository, evt: FileImportRequested): void {
	let input = document.createElement('input');
	input.className = 'hidden';
	input.type = 'file';
	input.id = evt.iD.value;
	input.multiple = evt.multiple.value;
	input.onchange = (event) => {
		const item = event.target as HTMLInputElement;
		if (!item.files) {
			return;
		}
		for (let i = 0; i < item.files.length; i++) {
			uploadRepository.fetchUpload(
				item.files[i],
				evt.iD.value,
				0,
				evt.scopeID.value,
				(uploauploadId: string, progress: number, total: number) => {
					console.log('progress', progress);
				},
				(uploadId) => {
				},
				(uploadId) => {
				},
				(uploadId) => {
					console.log('upload failed');
				}
			);
		}
	};
	if (evt.allowedMimeTypes.value) {
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

	chan.sendEvent(new RootViewAllocationRequested(
		getLocale(),
		evt.rootView,
		nextRID(),
		evt.values,
	));


	let url = `/${evt.rootView.value}`;
	if (evt.values.value && Object.entries(evt.values.value).length > 0) {
		url += '?';
		Object.entries(evt.values.value).forEach(([key, value], index, array) => {
			url += `${key}=${value}`;
			if (index < array.length - 1) {
				url += '&';
			}
		});
	}
	history.pushState(evt, '', url);
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
	if (evt.rootView && evt.rootView.value) {
		req.factory.value = evt.rootView.value;
	}

	if (req.factory.value === "") {
		req.factory.value = "."
	}

	if (evt.values && evt.values.value) {
		req.values = new RootViewParameters(evt.values.value);
	}

	req.locale = getLocale();
	req.rID = nextRID();

	chan.sendEvent(req);
}
