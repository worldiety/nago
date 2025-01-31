<script setup lang="ts">
import { ref } from 'vue';
import Magnifier from '@/assets/svg/magnifyingGlass.svg';

defineEmits<{
	(e: 'searchQueryChanged', searchQuery: string): void;
}>();

const searchQuery = ref<string>('');
</script>

<template>
	<div class="border-b-2 border-primary relative p-2.5 mx-2 my-2">
		<input
			v-model="searchQuery"
			@input="$emit('searchQueryChanged', searchQuery)"
			class="focus:outline-none w-full cursor-text pl-6"
			type="text"
			placeholder="Suchen nach..."
		/>
		<Magnifier class="absolute w-5 h-5 top-3.5 left-2" />
	</div>
</template>

<style scoped></style>
