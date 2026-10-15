/*
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

/**
 * A program runs synchronously and has no promises, so it loads an image as a handle, which it gets right away,
 * and learns about the result from an event, see package lib/asset. Drawing an image, which is not loaded, does
 * nothing. The images of a component are cached by their URL and released, when the component is unmounted.
 */

// The records of the events, see package lib/asset.
export const assetLoaded = 'go.wdy.de/nago/pkg/gostrict/lib/asset.Loaded';
export const assetFailed = 'go.wdy.de/nago/pkg/gostrict/lib/asset.Failed';

// maxEdge limits the size of a decoded image, e.g. a photo of a camera, so that it does not exhaust the memory of
// the GPU. A larger image is scaled down.
const maxEdge = 4096;

export interface ImageHandle {
	url: string;
	// bitmap is null, until the image is loaded
	bitmap: ImageBitmap | null;
	width: number;
	height: number;
	// users are told, when the image is loaded, e.g. a texture of three.js
	users: Array<(img: ImageHandle) => void>;
}

export type Emit = (type: string, values: Record<string, unknown>) => void;

/**
 * sameOrigin resolves the URL against the page and rejects any other origin.
 */
export function sameOrigin(url: string): URL {
	const abs = new URL(url, window.location.href);
	if (abs.origin !== window.location.origin) {
		throw new Error('only assets of the application may be loaded');
	}
	return abs;
}

export class AssetLoader {
	private readonly images = new Map<string, ImageHandle>();
	private closed = false;

	constructor(private readonly emit: Emit) {}

	/**
	 * image returns the handle of the image at the URL and loads it once. Only URLs of the origin of the page are
	 * allowed, e.g. the resources and uploads of nago, because a program must not send requests elsewhere. A
	 * texture of WebGL needs the image upside down, see flipY.
	 */
	image(url: string, flipY = false): ImageHandle {
		const key = flipY ? 'flip:' + url : url;
		let img = this.images.get(key);
		if (img) {
			if (img.bitmap) {
				// a program, which asks again, gets the event again
				this.emit(assetLoaded, { URL: url, Width: img.width, Height: img.height });
			}
			return img;
		}
		img = { url, bitmap: null, width: 0, height: 0, users: [] };
		this.images.set(key, img);
		this.load(img, flipY).catch((e) => {
			if (!this.closed) this.emit(assetFailed, { URL: url, Reason: String(e?.message ?? e) });
		});
		return img;
	}

	private async load(img: ImageHandle, flipY: boolean) {
		const res = await fetch(sameOrigin(img.url), { credentials: 'same-origin' });
		if (!res.ok) throw new Error(`HTTP ${res.status}`);
		const blob = await res.blob();
		// decoded off the main thread and rotated after its EXIF orientation, unless WebGL needs it upside down
		let bitmap = await createImageBitmap(blob, { imageOrientation: flipY ? 'flipY' : 'from-image' });
		const edge = Math.max(bitmap.width, bitmap.height);
		if (edge > maxEdge) {
			const scale = maxEdge / edge;
			const scaled = await createImageBitmap(bitmap, {
				resizeWidth: Math.round(bitmap.width * scale),
				resizeHeight: Math.round(bitmap.height * scale),
				resizeQuality: 'high',
			});
			bitmap.close();
			bitmap = scaled;
		}
		if (this.closed) {
			bitmap.close();
			return;
		}
		img.bitmap = bitmap;
		img.width = bitmap.width;
		img.height = bitmap.height;
		for (const u of img.users) u(img);
		this.emit(assetLoaded, { URL: img.url, Width: img.width, Height: img.height });
	}

	/**
	 * loaded reports a loaded asset, which another loader has loaded, e.g. a model.
	 */
	loaded(url: string, width = 0, height = 0) {
		if (!this.closed) this.emit(assetLoaded, { URL: url, Width: width, Height: height });
	}

	/**
	 * failed reports an asset, which can not be loaded.
	 */
	failed(url: string, e: unknown) {
		if (!this.closed) this.emit(assetFailed, { URL: url, Reason: String((e as Error)?.message ?? e) });
	}

	dispose() {
		this.closed = true;
		for (const img of this.images.values()) {
			img.bitmap?.close();
			img.bitmap = null;
		}
		this.images.clear();
	}
}
