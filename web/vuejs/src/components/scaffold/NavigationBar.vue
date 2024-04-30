<template>
	<nav class="fixed top-0 left-0 right-0 z-30">
		<div class="bg-white dark:bg-darkmode-gray border-b border-b-disabled-background dark:border-b-disabled-text h-16 py-2">
			<div class="website-content flex justify-between items-center h-full">
				<!-- TODO: Content here -->
				<div class="h-full *:h-full" v-html="ui.logo.v"></div>
				<ThemeToggle />
			</div>
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
