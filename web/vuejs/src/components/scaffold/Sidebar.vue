<template>
	<nav
		class="fixed top-0 left-0 bottom-0 z-10 h-full w-32 bg-white dark:bg-darkmode-gray"
		aria-label="Sidebar"
	>
		<div class="flex flex-col justify-between items-center h-full p-2">
			<div class="*:w-full" v-html="ui.logo.v"></div>
			<ThemeToggle />
		</div>
	</nav>
</template>

<script setup lang="ts">
import type { NavigationComponent } from '@/shared/protocol/gen/navigationComponent';
import ThemeToggle from '@/components/scaffold/ThemeToggle.vue';

defineProps<{
	ui: NavigationComponent;
}>();
</script>
