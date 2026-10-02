// Helpers which are injected into each page before a screenshot is taken. Rects are in document
// coordinates (CSS pixels), because the capture uses captureBeyondViewport.
(() => {
	if (window.__nagoShot) {
		return;
	}

	const roots = () => [
		...document.querySelectorAll('.content-container, #ora-overlay, #ora-modals, #ora-notifications'),
	];

	const visible = (el) => {
		const r = el.getBoundingClientRect();
		if (r.width < 1 || r.height < 1) {
			return false;
		}

		const s = getComputedStyle(el);
		return s.visibility !== 'hidden' && s.display !== 'none' && parseFloat(s.opacity) > 0.01;
	};

	const toDoc = (r) => ({ X: r.left + scrollX, Y: r.top + scrollY, W: r.width, H: r.height });

	const hasOwnText = (el) =>
		[...el.childNodes].some((n) => n.nodeType === Node.TEXT_NODE && n.textContent.trim() !== '');

	const media = new Set(['IMG', 'SVG', 'svg', 'CANVAS', 'INPUT', 'TEXTAREA', 'SELECT', 'BUTTON', 'VIDEO', 'IFRAME']);

	const transparent = (c) => c === 'transparent' || c === 'rgba(0, 0, 0, 0)';

	// paints tells whether an element draws something by itself: text, media or a box which differs from
	// the page background.
	const paints = (el, pageBg) => {
		if (hasOwnText(el) || media.has(el.tagName)) {
			return true;
		}

		const s = getComputedStyle(el);
		if (!transparent(s.backgroundColor) && s.backgroundColor !== pageBg) {
			return true;
		}

		if (s.boxShadow !== 'none') {
			return true;
		}

		return ['Top', 'Right', 'Bottom', 'Left'].some(
			(side) => parseFloat(s['border' + side + 'Width']) > 0 && !transparent(s['border' + side + 'Color'])
		);
	};

	const union = (rects) => {
		if (rects.length === 0) {
			return { X: 0, Y: 0, W: 0, H: 0 };
		}

		const x0 = Math.min(...rects.map((r) => r.X));
		const y0 = Math.min(...rects.map((r) => r.Y));
		const x1 = Math.max(...rects.map((r) => r.X + r.W));
		const y1 = Math.max(...rects.map((r) => r.Y + r.H));
		return { X: x0, Y: y0, W: x1 - x0, H: y1 - y0 };
	};

	const autoRect = () => {
		const container = document.querySelector('.content-container');
		const pageBg = container ? getComputedStyle(container).backgroundColor : '';
		const rects = [];
		for (const root of roots()) {
			for (const el of root.querySelectorAll('*')) {
				if (el.closest('svg') && el.tagName.toLowerCase() !== 'svg') {
					continue; // the svg itself is already accounted for
				}

				if (visible(el) && paints(el, pageBg)) {
					rects.push(toDoc(el.getBoundingClientRect()));
				}
			}
		}

		return union(rects);
	};

	const docSize = () => ({
		docW: Math.max(document.documentElement.scrollWidth, innerWidth),
		docH: Math.max(document.documentElement.scrollHeight, innerHeight),
	});

	const firstVisible = (selector) => [...document.querySelectorAll(selector)].find(visible);

	const center = (el) => {
		if (!el) {
			return { X: 0, Y: 0, OK: false };
		}

		el.scrollIntoView({ block: 'center', inline: 'center' });
		const r = el.getBoundingClientRect();
		return { X: r.left + r.width / 2, Y: r.top + r.height / 2, OK: true };
	};

	window.__nagoShot = {
		crop(mode) {
			let rect;
			if (mode === 'auto') {
				rect = autoRect();
			} else if (mode === 'viewport') {
				rect = { X: scrollX, Y: scrollY, W: innerWidth, H: innerHeight };
			} else if (mode === 'page') {
				const d = docSize();
				rect = { X: 0, Y: 0, W: d.docW, H: d.docH };
			} else {
				rect = union([...document.querySelectorAll(mode)].filter(visible).map((el) => toDoc(el.getBoundingClientRect())));
			}

			return { rect, ...docSize() };
		},

		centerOf(selector) {
			return center(firstVisible(selector));
		},

		// centerOfText finds the innermost visible element whose whole text equals the given text.
		centerOfText(text) {
			const candidates = [...document.querySelectorAll('body *')].filter(
				(el) => visible(el) && el.textContent.trim() === text
			);
			const innermost = candidates.filter((el) => !candidates.some((other) => other !== el && el.contains(other)));
			return center(innermost[0]);
		},

		async settle() {
			const deadline = Date.now() + 15000;
			while (Date.now() < deadline) {
				const c = document.querySelector('.content-container');
				if (c && (c.textContent.trim() !== '' || c.querySelector('img,svg,canvas'))) {
					break;
				}
				await new Promise((r) => setTimeout(r, 100));
			}

			await document.fonts.ready;
			await Promise.all(
				[...document.images].filter((img) => !img.complete).map((img) => new Promise((r) => (img.onload = img.onerror = r)))
			);

			// quiet period: no DOM mutation for 400ms, but give up after 5s for pages which animate forever
			await new Promise((resolve) => {
				let timer;
				const done = () => {
					obs.disconnect();
					resolve();
				};
				const obs = new MutationObserver(() => {
					clearTimeout(timer);
					timer = setTimeout(done, 400);
				});
				obs.observe(document.body, { subtree: true, childList: true, attributes: true, characterData: true });
				timer = setTimeout(done, 400);
				setTimeout(done, 5000);
			});

			const c = document.querySelector('.content-container');
			return !!c && c.children.length > 0;
		},
	};
})();
