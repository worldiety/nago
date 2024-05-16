<template>
	<div
		class="menu-entry flex justify-start items-center gap-x-2 cursor-pointer hover:bg-disabled-background hover:bg-opacity-25 active:bg-opacity-35 rounded-full w-full p-4"
		tabindex="0"
		@click="expandMenuEntry"
		@keydown.enter="expandMenuEntry"
	>
		<div class="relative h-full">
			<div class="menu-entry-icon h-4 *:h-full" v-html="ui.icon.v"></div>
			<div class="menu-entry-icon-active h-4 *:h-full" v-html="ui.iconActive.v"></div>
			<!-- Optional red badge -->
			<div
				v-if="ui.badge.v"
				class="absolute -top-1.5 -right-1.5 flex justify-center items-center h-3.5 px-1 rounded-full bg-error"
			>
				<p class="text-xs text-white">{{ ui.badge.v }}</p>
			</div>
		</div>
		<p class="grow leading-tight select-none">{{ ui.title.v }}</p>
		<TriangleDown v-if="hasSubMenuEntries" class="shrink-0 basis-2" :class="{'rotate-180': ui.expanded.v}" />
	</div>
</template>

<script setup lang="ts">
import TriangleDown from '@/assets/svg/triangleDown.svg';
import type { MenuEntry } from '@/shared/protocol/ora/menuEntry';
import { computed } from 'vue';
import { useServiceAdapter } from '@/composables/serviceAdapter';

const props = defineProps<{
	ui: MenuEntry;
}>();

const serviceAdapter = useServiceAdapter();

const hasSubMenuEntries = computed((): boolean => {
	return props.ui.menu.v && props.ui.menu.v.length > 0;
});

function expandMenuEntry(): void {
	if (hasSubMenuEntries.value) {
		serviceAdapter.setPropertiesAndCallFunctions([
			{
				...props.ui.expanded,
				v: !props.ui.expanded.v,
			},
		], [props.ui.onFocus]);
	}
}
</script>

<style scoped>
.menu-entry:hover .menu-entry-icon,
.menu-entry .menu-entry-icon-active {
	@apply hidden;
}

.menu-entry:hover .menu-entry-icon-active {
	@apply block;
}
</style>
