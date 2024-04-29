<template>
	<nav class="fixed top-0 left-0 right-0 flex justify-between items-center bg-white dark:bg-darkmode-gray border-b border-b-disabled-background dark:border-b-disabled-text h-16 p-2 z-30">
		<!-- TODO: Content here -->
		<div class="h-full *:h-full" v-html="ui.logo.v"></div>
		<ThemeToggle />
	</nav>
</template>

<script setup lang="ts">
import type { NavigationComponent } from '@/shared/protocol/gen/navigationComponent';
import ThemeToggle from '@/components/scaffold/ThemeToggle.vue';

defineProps<{
	ui: NavigationComponent;
}>();
</script>
