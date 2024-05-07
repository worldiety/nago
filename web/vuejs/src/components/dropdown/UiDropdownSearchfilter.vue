<script setup lang="ts">
import Magnifier from '@/assets/svg/magnifyingGlass.svg'
import {ref} from "vue";

defineEmits<{
	(e: 'searchQueryChanged', searchQuery: string): void
}>()

const searchQuery = ref<string>('')



</script>

<template>
	<div class="border-b-2 border-ora-orange relative p-2.5 mx-1">
		<input v-model="searchQuery" @input="$emit('searchQueryChanged', searchQuery)" class="outline-none w-full cursor-text pl-6 dark:bg-ora-dropdown-background dark:text-white" type="text" placeholder="Suchen nach...">
		<Magnifier class="absolute w-5 h-5 top-3.5 left-2 dark:text-white"/>
	</div>
</template>

<style scoped>

</style>
