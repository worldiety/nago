<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script setup lang="ts">
import { computed } from 'vue';
import Editor from '@/components/richtexteditor/Editor.vue';
import { frameCSSObject } from '@/components/shared/frame';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { RichTextEditor, UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: RichTextEditor;
}>();

const serviceAdapter = useServiceAdapter();

const frameStyles = computed<Object | undefined>(() => {
	let styles = frameCSSObject(props.ui.frame);

	return styles;
});

// security note: an attacker may have introduced malicious html into the model,
// however, tiptap removes everything it does not know.

// security note: an attacker may submit malicious html through the protocol itself.
// We cannot do anything against that from the frontend side and the backend must
// sanitize that.

let lastSent: string | undefined;

function onBlur(): any {
	//console.log('blurred', props.ui.value);
	if (props.ui.inputValue && lastSent !== props.ui.value) {
		serviceAdapter.sendEvent(
			new UpdateStateValueRequested(props.ui.inputValue, undefined, nextRID(), props.ui.value)
		);

		lastSent = props.ui.value;
	}
	return undefined;
}
</script>

<template v-if="props.ui.iv">
	<editor v-model="ui.value" :style="frameStyles" :disabled="props.ui.disabled" @blur="onBlur" />
</template>
