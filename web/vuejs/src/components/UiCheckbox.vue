<script setup lang="ts">
import { ref, watch } from 'vue';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import type { Checkbox } from '@/shared/protocol/ora/checkbox';

const props = defineProps<{
	ui: Checkbox;
}>();

const serviceAdapter = useServiceAdapter();

const checked = ref<boolean>(props.ui.v ? props.ui.v : false);

watch(
	() => props.ui.v,
	(newValue) => {
		if (newValue) {
			checked.value = newValue;
		} else {
			checked.value = false;
		}
	}
);

function checkboxSelected(): void {
	if (!props.ui.d) {
		serviceAdapter.setProperties({
			p: props.ui.i,
			v: !checked.value,
		});
	}
}
</script>

<template>
	<div
		v-if="!ui.iv"
		class="input-checkbox rounded-full w-fit"
		:class="{ 'input-checkbox-disabled': ui.d }"
		:tabindex="ui.d ? '-1' : '0'"
		@click="checkboxSelected"
		@keydown.enter="checkboxSelected"
	>
		<div class="p-2.5">
			<input :checked="checked" type="checkbox" class="pointer-events-none" tabindex="-1" :disabled="ui.d" />
		</div>
	</div>
</template>

<style scoped>
.input-checkbox:hover {
	@apply bg-I0 bg-opacity-25;
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
	@apply border-I0;
}

.input-checkbox.input-checkbox-disabled:hover input:not(:checked) {
	@apply border-ST0;
}
</style>
