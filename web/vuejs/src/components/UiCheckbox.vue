<script setup lang="ts">
import { ref } from 'vue';
import type {Checkbox} from "@/shared/protocol/ora/checkbox";
import {useServiceAdapter} from "@/composables/serviceAdapter";

const props = defineProps<{
	ui: Checkbox
}>()

const serviceAdapter = useServiceAdapter()

function checkboxSelected(): void {
	if (!props.ui.disabled.v) {
		serviceAdapter.setPropertiesAndCallFunctions([{
			...props.ui.selected, v: !props.ui.selected.v
		}], [props.ui.clicked])
	}
}
</script>

<template>
	<div
		class="input-checkbox rounded-full w-fit -ml-2.5"
		:class="{'input-checkbox-disabled': ui.disabled.v}"
		:tabindex="ui.disabled.v ? '-1' : '0'"
		@click="checkboxSelected"
		@keydown.enter="checkboxSelected"
	>
		<div class="p-2.5">
			<input :checked="ui.selected.v" type="checkbox" class="pointer-events-none" tabindex="-1" :disabled="ui.disabled.v">
		</div>
	</div>
</template>


<style scoped>
.input-checkbox:hover {
	@apply bg-ora-orange bg-opacity-25;
}

.input-checkbox:active {
	@apply bg-opacity-35;
}

.input-checkbox:focus-visible {
	@apply outline-none outline-black outline-offset-2 ring-white ring-2;
}

.input-checkbox.input-checkbox-disabled:hover {
	@apply bg-transparent;
}

.input-checkbox.input-checkbox-disabled:focus-visible {
	@apply outline-none ring-0;
}

.input-checkbox:hover input:not(:checked) {
	@apply border-ora-orange;
}

.input-checkbox.input-checkbox-disabled:hover input:not(:checked) {
	@apply border-disabled-text;
}
</style>

