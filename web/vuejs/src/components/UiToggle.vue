<script lang="ts" setup>
import { ref, watch } from 'vue';
import type {Toggle} from "@/shared/protocol/ora/toggle";
import { useServiceAdapter } from '@/composables/serviceAdapter';

const props = defineProps<{
	ui: Toggle;
}>();

const serviceAdapter = useServiceAdapter();
const checked = ref<boolean>(props.ui.checked.v);

watch(() => props.ui.checked.v, (newValue) => {
	checked.value = newValue;
})

function onClick() {
	if (props.ui.disabled.v) {
		return;
	}
	checked.value = !checked.value;
	serviceAdapter.setPropertiesAndCallFunctions([{
		...props.ui.checked,
		v: checked.value,
	}], [props.ui.onCheckedChanged]);
}
</script>

<template>
	<div>
		<span v-if="props.ui.label.v" class="block mb-2 text-sm">{{ props.ui.label.v }}</span>
		<div
			class="toggle-switch-container"
			:class="{'toggle-switch-container-disabled': props.ui.disabled.v}"
			:tabindex="props.ui.disabled.v ? '-1' : '0'"
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
	@apply relative h-6 w-11 rounded-full outline outline-1 outline-black;
	@apply dark:outline-white dark:after:outline-white;
}

.toggle-switch::after {
	@apply absolute start-[6px] top-1 h-4 w-4 rounded-full border border-black bg-transparent transition-transform content-[''];
	@apply dark:border-white;
}

.toggle-switch.toggle-switch-checked {
	@apply after:translate-x-[105%] after:border-ora-orange after:bg-ora-orange;
}

.toggle-switch-container {
	@apply inline-block rounded-full p-1.5 -ml-1.5;
}

.toggle-switch-container:hover {
	@apply bg-ora-orange bg-opacity-25;
}

.toggle-switch-container:active {
	@apply bg-opacity-35;
}

.toggle-switch-container:hover .toggle-switch {
	@apply outline-ora-orange;
}

.toggle-switch-container:hover .toggle-switch::after {
	@apply border-ora-orange;
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
	@apply outline-disabled-text;
}

.toggle-switch-container.toggle-switch-container-disabled .toggle-switch::after {
	@apply bg-transparent border-disabled-text;
}

.toggle-switch-container.toggle-switch-container-disabled .toggle-switch.toggle-switch-checked::after {
	@apply bg-disabled-text;
}
</style>
