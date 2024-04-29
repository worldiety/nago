<template>
	<div class="fixed z-30 flex w-full flex-1 flex-col dark:text-white">
		<nav class="flex h-16 justify-between bg-white px-4 shadow dark:bg-gray-700">
			<!-- TODO: Content here -->
		</nav>
	</div>
</template>

<script setup lang="ts">
import type { NavigationComponent } from '@/shared/protocol/gen/navigationComponent';

defineProps<{
	ui: NavigationComponent;
}>();
</script>
