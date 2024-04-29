<template>
	<nav
		class="fixed left-0 z-20 h-screen w-64 -translate-x-full pt-16 transition-transform sm:translate-x-0"
		aria-label="Sidebar"
	>
		<div class="h-full overflow-y-auto bg-gray-50 px-3 py-4 dark:bg-gray-800">
			<!-- TODO: Content here -->
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
