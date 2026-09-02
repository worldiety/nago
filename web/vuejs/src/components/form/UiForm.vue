<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script setup lang="ts">
import { computed, provide } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { frameCSS } from '@/components/shared/frame';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { Form, FunctionCallRequested } from '@/shared/proto/nprotoc_gen';
import { PRE_SUBMIT_GROUP, PreSubmitGroup } from '@/components/form/preSubmitGroup';

const props = defineProps<{
	ui: Form;
}>();

const serviceAdapter = useServiceAdapter();

const preSubmitDelay = 50; // delay in ms to wait for pre submit group members
let submitTimeout: any = null;

const preSubmitGroup = new PreSubmitGroup();
provide(PRE_SUBMIT_GROUP, preSubmitGroup);

function handleSubmit(event: Event) {
	// tell members of the pre submit group (inputs) to send their current values to the backend before the form is submitted
	preSubmitGroup.execute();

	if (submitTimeout) {
		clearTimeout(submitTimeout);
		submitTimeout = null;
	}

	// Use delay as workaround to allow all input fields of the form to update their values before sending the submit event.
	submitTimeout = setTimeout(() => {
		if (props.ui.action) {
			event.stopPropagation();
			serviceAdapter.sendEvent(new FunctionCallRequested(props.ui.action, nextRID()));
		}
	}, preSubmitDelay);
}

const frameStyles = computed<string>(() => {
	let styles = frameCSS(props.ui.frame);

	return styles.join(';');
});
</script>

<template v-if="props.ui.iv">
	<form
		:id="ui.id"
		:style="frameStyles"
		:autocomplete="ui.autocomplete ? 'on' : 'off'"
		@submit.prevent="handleSubmit"
	>
		<ui-generic v-for="childUi in props.ui.children?.value" :ui="childUi" />
		<button type="submit" class="hidden" tabindex="-1" aria-hidden="true">
			<!-- hidden submit button to enable enter key submission -->
		</button>
	</form>
</template>
