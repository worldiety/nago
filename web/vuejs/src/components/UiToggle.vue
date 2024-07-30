<script lang="ts" setup>
import {ref, watch} from 'vue';
import type {Toggle} from "@/shared/protocol/ora/toggle";
import {useServiceAdapter} from '@/composables/serviceAdapter';

const props = defineProps<{
	ui: Toggle;
}>();

const serviceAdapter = useServiceAdapter();
const checked = ref<boolean>(props.ui.v ? props.ui.v : false);

watch(() => props.ui.v, (newValue) => {
	if (newValue) {
		checked.value = newValue;
	} else {
		checked.value = false;
	}
})

function onClick() {
	if (props.ui.d) {
		return;
	}

	serviceAdapter.setProperties({
		p: props.ui.i,
		v: !checked.value,
	});
}
</script>

<template>
	<div v-if="!ui.iv">
		<div
			class="toggle-switch-container"
			:class="{'toggle-switch-container-disabled': props.ui.d}"
			:tabindex="props.ui.d ? '-1' : '0'"
			@click="onClick"
			@keydown.enter="onClick"
		>
			<div
				class="toggle-switch"
				:class="{'toggle-switch-checked': checked}"
			></div>
		</div>
	</div>
</template>

<style scoped>
.toggle-switch {
	@apply relative h-6 w-11 rounded-full outline outline-1;
}

.toggle-switch::after {
	@apply absolute start-[6px] top-1 h-4 w-4 rounded-full border bg-transparent transition-transform content-[''];
}

.toggle-switch.toggle-switch-checked {
	@apply after:translate-x-[105%] after:border-I0 after:bg-I0;
}

.toggle-switch-container {
	@apply inline-block rounded-full;
}

.toggle-switch-container:hover {
	@apply bg-I0 bg-opacity-25;
}

.toggle-switch-container:active {
	@apply bg-opacity-35;
}

.toggle-switch-container:hover .toggle-switch {
	@apply outline-I0;
}

.toggle-switch-container:hover .toggle-switch::after {
	@apply border-I0;
}

.toggle-switch-container:focus-visible {
	@apply outline-none outline-2 outline-offset-2 outline-black ring-white ring-2;
}

.toggle-switch-container.toggle-switch-container-disabled:hover {
	@apply bg-transparent;
}

.toggle-switch-container.toggle-switch-container-disabled:focus-visible {
	@apply outline-none ring-0;
}

.toggle-switch-container.toggle-switch-container-disabled .toggle-switch {
	@apply outline-ST0;
}

.toggle-switch-container.toggle-switch-container-disabled .toggle-switch::after {
	@apply bg-transparent border-ST0;
}

.toggle-switch-container.toggle-switch-container-disabled .toggle-switch.toggle-switch-checked::after {
	@apply bg-ST0;
}
</style>
