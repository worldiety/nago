<template>
	<div class="fixed top-0 left-0 right-0 text-black dark:text-white bg-white dark:bg-darkmode-gray border-b border-b-disabled-background dark:border-b-disabled-text h-24 p-4 z-30">
		<div class="relative flex justify-start items-center h-full">
			<MenuIcon class="h-6" />
			<div class="absolute top-0 left-0 bottom-0 right-0 flex justify-center items-center h-full">
				<div class="h-full *:h-full" v-html="ui.logo.v"></div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import MenuIcon from '@/assets/svg/menu.svg';
import type { NavigationComponent } from '@/shared/protocol/ora/navigationComponent';

defineProps<{
	ui: NavigationComponent;
}>();
</script>
