type Callback = (err: Error | unknown | null, data: unknown) => void;

// this script loads either a legacy (polyfilled) or modern build based on the client's browser version,
// using Bowser to detect the browser and comparing it against version thresholds.
(function () {
	function loadJSON(url: string, callback: Callback) {
		// we do not use let or const in this script to perform the script on legacy browsers as well
		const xhr = new XMLHttpRequest();
		// we do not use fetch here to perform the request on legacy browsers as well,
		xhr.open('GET', url, true);
		xhr.onreadystatechange = function () {
			// ready state 4 is 'DONE'
			if (xhr.readyState === 4) {
				if (xhr.status >= 200 && xhr.status < 300) {
					try {
						const data = JSON.parse(xhr.responseText);
						callback(null, data);
					} catch (e) {
						callback(e, null);
					}
				} else {
					callback(new Error('Could not load URL: ' + url), null);
				}
			}
		};
		try {
			xhr.send(null);
		} catch (e) {
			callback(e, null);
		}
	}
	function loadScript(src: string, type: string) {
		const script = document.createElement('script');
		script.src = src;
		script.defer = true;
		script.type = type;
		document.head.appendChild(script);
		console.debug(`Loaded script '${src}' with type '${type}'`);
	}
	function loadLink(rel: string, href: string) {
		const link = document.createElement('link');
		link.rel = rel;
		link.href = href;
		document.head.appendChild(link);
		console.debug(`Loaded link '${rel}' with href '${href}'`);
	}

	// eslint-disable-next-line
	const browser = (window as any).bowser.getParser(window.navigator.userAgent);
	const info = browser.getBrowser();
	const name = info.name;
	const version = parseInt(info.version, 10);
	const thresholds: Record<string, number> = {
		'Chrome': 103,
		'Firefox': 102,
		'Edge': 103,
		'Safari': 15,
		'IE': Infinity,
		'Internet Explorer': Infinity,
		'Opera': 89,
	};

	const isOutdated = thresholds[name] !== undefined && version < thresholds[name];
	const buildType = isOutdated ? 'legacy' : 'modern';

	console.info(`Detected browser '${name}' with version '${version}'. Loading build '${buildType}'...`);

	try {
		loadJSON(`/${buildType}/manifest.json`, function (err, manifest) {
			if (err) {
				console.error(`Error while loading build manifest for build ${buildType}`, err);
				return;
			}

			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			const mainEntry = (manifest as any)['src/main.ts'];
			if (!mainEntry || !mainEntry.file) {
				console.error(`No entry file in build manifest found for build ${buildType}`);
				return;
			}

			if (isOutdated) {
				// manually load polyfills
				loadScript(`/${buildType}/assets/polyfill.min.js`, 'text/javascript');
				loadScript(`/${buildType}/assets/minified.js`, 'text/javascript');
				loadScript(`/${buildType}/assets/runtime.js`, 'text/javascript');
			}

			// load build and stylesheets
			loadScript(`/${buildType}/${mainEntry.file}`, isOutdated ? 'text/javascript' : 'module');
			if (mainEntry.css) {
				// eslint-disable-next-line @typescript-eslint/no-explicit-any
				mainEntry.css.forEach((cssFile: any) => {
					loadLink('stylesheet', `/${buildType}/${cssFile}`);
				});
			}

			loadLink('manifest', `/${buildType}/manifest.json`);
			console.info(`Successfully loaded build '${buildType}' for browser '${name}'.
Version '${version}' is ${isOutdated ? 'older' : 'newer'} than the threshold '${thresholds[name]}'.`);
		});
	} catch (err) {
		console.error('Error while initializing the build:', err);
	}
})();
