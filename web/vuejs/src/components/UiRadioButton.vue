<script setup lang="ts">
import { ref } from 'vue';
import type {Radiobutton} from "@/shared/protocol/ora/radiobutton";
import {useServiceAdapter} from "@/composables/serviceAdapter";

const props = defineProps<{
	ui: Radiobutton
}>();

const serviceAdapter = useServiceAdapter()

function radioButtonClicked(): void {
	if (!props.ui.disabled.v) {
	serviceAdapter.setPropertiesAndCallFunctions([{
		...props.ui.selected, v: !props.ui.selected.v
	}], [props.ui.onClicked])
	}
}
</script>

<template>
	<div
		v-if="ui.visible.v"
		class="input-radio rounded-full w-fit -ml-2.5"
		:class="{'input-radio-disabled': ui.disabled.v}"
		:tabindex="ui.disabled.v ? '-1' : '0'"
		@click="radioButtonClicked"
		@keydown.enter="radioButtonClicked"
	>
		<div class="p-2.5">
			<input :checked="ui.selected.v" type="radio" class="pointer-events-none" tabindex="-1" :disabled="ui.disabled.v">
		</div>
	</div>
</template>

<style scoped>
.input-radio:hover {
	@apply bg-primary bg-opacity-25;
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
	@apply border-primary;
}

.input-radio.input-radio-disabled:hover input:checked {
	@apply border-disabled-text;
}
</style>
