<script setup lang="ts">
import {ref, watch} from 'vue';
import {useServiceAdapter} from '@/composables/serviceAdapter';
import {Ptr, Radiobutton, Str, UpdateStateValueRequested} from "@/shared/proto/nprotoc_gen";
import {nextRID} from "@/eventhandling";
import {bool2Str} from "@/components/shared/util";

const props = defineProps<{
	ui: Radiobutton;
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

function radioButtonClicked(): void {
	if (!props.ui.disabled.value) {
		serviceAdapter.sendEvent(new UpdateStateValueRequested(
			props.ui.inputValue,
			new Ptr(),
			nextRID(),
			new Str(bool2Str(!checked.value)),
		))
	}
}
</script>

<template>
	<div
		v-if="!ui.invisible.value"
		class="input-radio rounded-full w-fit"
		:class="{ 'input-radio-disabled': ui.disabled.value }"
		:tabindex="ui.disabled.value ? '-1' : '0'"
		@click="radioButtonClicked"
		@keydown.enter="radioButtonClicked"
	>
		<div class="p-2.5">
			<input :checked="checked" type="radio" class="pointer-events-none" :disabled="ui.disabled.value"/>
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
