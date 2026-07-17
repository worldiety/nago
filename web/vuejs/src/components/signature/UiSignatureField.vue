<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->
<template>
	<div class="signature-field" :style="frameStyles">
		<SignatureInput :ui="props.ui" @expand="expanded = true" />
		<SignatureOverlay v-if="expanded" :ui="props.ui" class="overlay" @submit="onSubmit" @close="expanded = false"/>
	</div>
</template>
<script lang="ts" setup>
import { computed, ref } from 'vue';
import { frameCSS } from '@/components/shared/frame';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { Signature, SignatureField, UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';
import SignatureOverlay from '@/components/signature/SignatureOverlay.vue';
import SignatureInput from '@/components/signature/SignatureInput.vue';

const props = defineProps<{
	ui: SignatureField;
}>();

const serviceAdapter = useServiceAdapter();
const expanded = ref(false);

const frameStyles = computed<string>(() => {
	const styles = frameCSS(props.ui.frame);

	return styles.join(';');
});

function onSubmit(signature: Signature): void {
	serviceAdapter.sendEvent(
		new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), JSON.stringify(signature))
	);
	expanded.value = false;
}
</script>
