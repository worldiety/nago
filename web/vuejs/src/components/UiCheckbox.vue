<script setup lang="ts">
import { ref, watch } from 'vue';
import { bool2Str } from '@/components/shared/util';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { Checkbox, Ptr, Str, UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Checkbox;
}>();

const serviceAdapter = useServiceAdapter();

const checked = ref<boolean>(props.ui.value.value);

watch(
	() => props.ui.value.value,
	(newValue) => {
		if (newValue) {
			checked.value = newValue;
		} else {
			checked.value = false;
		}
	}
);

function checkboxSelected(): void {
	if (!props.ui.disabled.value) {
		serviceAdapter.sendEvent(
			new UpdateStateValueRequested(props.ui.inputValue, new Ptr(), nextRID(), new Str(bool2Str(!checked.value)))
		);
	}
}
</script>

<template>
	<div
		v-if="!ui.invisible.value"
		class="input-checkbox rounded-full w-fit"
		:class="{ 'input-checkbox-disabled': ui.disabled.value }"
		:tabindex="ui.disabled.value ? '-1' : '0'"
		@click="checkboxSelected"
		@keydown.enter="checkboxSelected"
	>
		<div class="p-2.5">
			<input
				:checked="checked"
				type="checkbox"
				class="pointer-events-none"
				tabindex="-1"
				:disabled="ui.disabled.value"
			/>
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
