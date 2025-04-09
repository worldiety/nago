<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script setup lang="ts">
import { computed } from 'vue';
import { frameCSS } from '@/components/shared/frame';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import sanitizeHtml from 'sanitize-html';
import { RichText } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: RichText;
}>();

const serviceAdapter = useServiceAdapter();

// security note: an attacker may inject malicious html into our model and the server or backend developer
// forgot to check or remove it.
// Thus, let us sanitize the input.

const sanitizedHtml = computed<string>(() => {
	if (!props.ui.value) {
		return '';
	}

	return sanitizeHtml(props.ui.value);
});

const frameStyles = computed<string>(() => {
	let styles = frameCSS(props.ui.frame);
	if (!styles) {
		return '';
	}

	return styles.join(';');
});
</script>

<template>
	<div v-html="sanitizedHtml" :style="frameStyles" class="prose-custom" />
</template>
