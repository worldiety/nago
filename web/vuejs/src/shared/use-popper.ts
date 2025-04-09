/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { onMounted, ref, watchEffect } from 'vue';
import { createPopper } from '@popperjs/core';

export function usePopper(options) {
	let reference = ref<HTMLElement | null>(null);
	let popper = ref<HTMLElement | null>(null);

	onMounted(() => {
		watchEffect((onInvalidate) => {
			if (!popper.value) return;
			if (!reference.value) return;

			let popperEl = popper.value.el || popper.value;
			let referenceEl = reference.value.el || reference.value;

			if (!(referenceEl instanceof HTMLElement)) return;
			if (!(popperEl instanceof HTMLElement)) return;

			let { destroy } = createPopper(referenceEl, popperEl, options);

			onInvalidate(destroy);
		});
	});

	return [reference, popper];
}
