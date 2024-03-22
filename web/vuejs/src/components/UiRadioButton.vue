<script setup lang="ts">
import { ref } from 'vue';

const props = defineProps<{
	disabled?: boolean;
}>();

const checked = ref<boolean>(false);

function radioButtonClicked(): void {
	if (!props.disabled) {
		checked.value = !checked.value;
	}
}
</script>

<template>
	<div
		class="input-radio rounded-full w-fit -ml-2.5"
		:class="{'input-radio-disabled': disabled}"
		:tabindex="disabled ? '-1' : '0'"
		@click="radioButtonClicked"
		@keydown.enter="radioButtonClicked"
	>
		<div class="p-2.5">
			<input :checked="checked" type="radio" class="pointer-events-none" tabindex="-1" :disabled="disabled">
		</div>
	</div>
</template>

<style scoped>
.input-radio:hover {
	@apply bg-ora-orange bg-opacity-25;
}

.input-radio:active {
	@apply bg-opacity-35;
}

.input-radio:focus {
	@apply outline-none outline-black outline-offset-2 ring-white ring-2;
}

.input-radio.input-radio-disabled:hover {
	@apply bg-transparent;
}

.input-radio.input-radio-disabled:focus {
	@apply outline-none ring-0;
}

.input-radio:hover input:checked {
	@apply border-ora-orange;
}

.input-radio.input-radio-disabled:hover input:checked {
	@apply border-disabled-text;
}
</style>
