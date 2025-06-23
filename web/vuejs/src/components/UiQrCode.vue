<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { frameCSS } from '@/components/shared/frame';
import { toDataURL } from 'qrcode';
import type { QrCode } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: QrCode;
}>();

const qrCodeUrl = ref<string | null>(null);

async function generateQrCode() {
	if (!props.ui.value) {
		qrCodeUrl.value = null;
		return;
	}

	qrCodeUrl.value = await toDataURL(props.ui.value, { type: 'image/webp' });
}

const frameStyles = computed<string>(() => {
	const styles = frameCSS(props.ui.frame);

	return styles.join(';');
});

onMounted(async () => {
	await generateQrCode();
});
</script>

<template>
	<div v-if="qrCodeUrl" :style="frameStyles">
		<img :src="qrCodeUrl" :alt="props.ui.accessibilityLabel" class="size-full" />
	</div>
</template>
