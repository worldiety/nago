<script setup lang="ts">
import { ref } from 'vue';

const props = defineProps<{
	disabled?: boolean;
}>()

const checked = ref<boolean>(false);

function checkboxClicked(): void {
	if (!props.disabled) {
		checked.value = !checked.value;
	}
}
</script>

<template>
	<div
		class="input-checkbox rounded-full w-fit -ml-2.5"
		:class="{'input-checkbox-disabled': disabled}"
		:tabindex="disabled ? '-1' : '0'"
		@click="checkboxClicked"
		@keydown.enter="checkboxClicked"
	>
		<div class="p-2.5">
			<input v-model="checked" type="checkbox" class="pointer-events-none" tabindex="-1" :disabled="disabled">
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

.input-checkbox:focus {
	@apply outline-none outline-black outline-offset-2 ring-white ring-2;
}

.input-checkbox.input-checkbox-disabled:hover {
	@apply bg-transparent;
}

.input-checkbox.input-checkbox-disabled:focus {
	@apply outline-none ring-0;
}

.input-checkbox:hover input:not(:checked) {
	@apply border-ora-orange;
}

.input-checkbox.input-checkbox-disabled:hover input:not(:checked) {
	@apply border-disabled-text;
}
</style>
