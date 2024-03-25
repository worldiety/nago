<script lang="ts" setup>
import { useNetworkStore } from '@/stores/networkStore';
import type { LiveToggle } from '@/shared/model/liveToggle';
import type { LivePage } from '@/shared/model/livePage';
import { ref } from 'vue';

const props = defineProps<{
	ui: LiveToggle;
	page: LivePage;
}>();

const networkStore = useNetworkStore();
const checked = ref<boolean>(props.ui.checked.value);

function onClick() {
	checked.value = !checked.value;
	networkStore.invokeFuncAndSetProp({
		...props.ui.checked,
		value: checked.value,
	}, props.ui.onCheckedChanged);
}
</script>

<template>
	<div>
		<span v-if="props.ui.label.value" class="block mb-2 text-sm font-medium">{{ props.ui.label.value }}</span>
		<div
			class="toggle-switch-container"
			tabindex="0"
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
	@apply absolute start-[6px] top-1 h-4 w-4 rounded-full border border-black bg-transparent transition-all content-[''];
}

.toggle-switch.toggle-switch-checked {
	@apply after:translate-x-[105%] after:border-ora-orange after:bg-ora-orange;
}

.toggle-switch.toggle-switch-disabled {

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

.toggle-switch-container:focus {
	@apply outline-none outline-2 outline-offset-2 outline-black ring-white ring-2;
}
</style>
