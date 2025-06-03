/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

export type UploadProgressItem = {
	id: string;
	fileName: string;
	progress: number;
	total: number;
};

type Subscriber = (items: UploadProgressItem[]) => void;

const uploads: UploadProgressItem[] = [];
const subscribers = new Set<Subscriber>();

export const uploadProgressManager = {
	addUpload(id: string, fileName: string, total: number) {
		uploads.push({ id, fileName, progress: 0, total: total });
		this.notify();
	},
	updateProgress(id: string, progress: number) {
		const item = uploads.find((u) => u.id === id);
		if (item) {
			item.progress = progress;
			this.notify();
		}
	},
	removeUpload(id: string) {
		const index = uploads.findIndex((u) => u.id === id);
		if (index !== -1) {
			uploads.splice(index, 1);
			this.notify();
		}
	},
	subscribe(fn: Subscriber) {
		subscribers.add(fn);
		fn(uploads); // initial push
		return () => subscribers.delete(fn);
	},
	notify() {
		for (const fn of subscribers) {
			fn([...uploads]);
		}
	},
};
