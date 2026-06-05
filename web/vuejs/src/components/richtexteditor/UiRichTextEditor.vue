<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { frameCSSObject } from '@/components/shared/frame';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { RichTextEditor, UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';
import Editor from '@/components/richtexteditor/Editor.vue';

const props = defineProps<{
	ui: RichTextEditor;
}>();

const serviceAdapter = useServiceAdapter();

const value = ref(props.ui.value);

const frameStyles = computed<object | undefined>(() => {
	return frameCSSObject(props.ui.frame);
});

// security note: an attacker may have introduced malicious html into the model,
// however, tiptap removes everything it does not know.

// security note: an attacker may submit malicious html through the protocol itself.
// We cannot do anything against that from the frontend side and the backend must
// sanitize that.

let lastSent: string | undefined;

function onBlur(): any {
	if (props.ui.inputValue && lastSent !== value.value) {
		serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.inputValue, undefined, nextRID(), value.value));

		lastSent = value.value;
	}
	return undefined;
}

watch(
	() => props.ui.value,
	() => {
		value.value = props.ui.value;
	}
);
</script>

<template v-if="props.ui.iv">
	<Editor v-model="value" :style="frameStyles" :disabled="props.ui.disabled" @blur="onBlur" />
</template>
