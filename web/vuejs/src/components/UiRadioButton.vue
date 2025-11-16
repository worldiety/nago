<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script setup lang="ts">
import { ref, watch } from 'vue';
import { bool2Str } from '@/components/shared/util';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { Radiobutton, UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Radiobutton;
}>();

const serviceAdapter = useServiceAdapter();

const checked = ref<boolean>(props.ui.value ? props.ui.value : false);

watch(
	() => props.ui.value,
	(newValue) => {
		if (newValue) {
			checked.value = newValue;
		} else {
			checked.value = false;
		}
	}
);

function radioButtonClicked(): void {
	if (!props.ui.disabled) {
		serviceAdapter.sendEvent(
			new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), bool2Str(!checked.value))
		);
	}
}
</script>

<template>
	<div
		v-if="!ui.invisible"
		class="input-radio rounded-full w-fit"
		:class="{ 'input-radio-disabled': ui.disabled }"
		:tabindex="ui.disabled ? '-1' : '0'"
		@click="radioButtonClicked"
		@keydown.enter="radioButtonClicked"
	>
		<div class="p-2.5">
			<input :id="ui.id" :checked="checked" type="radio" class="pointer-events-none" :disabled="ui.disabled" />
		</div>
	</div>
</template>

<style scoped>
.input-radio:hover {
	@apply bg-I0 bg-opacity-25;
}

.input-radio:active {
	@apply bg-opacity-35;
}

.input-radio:focus-visible {
	@apply outline-none outline-black outline-offset-2 ring-white ring-2;
}

.input-radio.input-radio-disabled:hover {
	@apply bg-transparent;
}

.input-radio.input-radio-disabled:focus-visible {
	@apply outline-none ring-0;
}

.input-radio:hover input:not(:disabled) {
	@apply border-I0;
}

.input-radio.input-radio-disabled:hover input:checked {
	@apply border-ST0;
}
</style>
