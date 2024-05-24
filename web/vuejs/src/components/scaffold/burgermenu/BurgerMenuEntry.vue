<template>
	<div class="flex flex-col justify-start items-start gap-y-4 w-full">
		<div
			class="menu-entry flex justify-start items-center gap-x-4 rounded-full w-full p-4"
			:class="{
				'cursor-pointer hover:bg-disabled-background hover:bg-opacity-25 active:bg-opacity-35': menuEntryClickable,
				'bg-disabled-background bg-opacity-35': menuEntryActive,
			}"
			:tabindex="menuEntryClickable ? '0' : '-1'"
			@click="menuEntryClicked"
			@keydown.enter="menuEntryClicked"
		>
			<div v-if="ui.icon.v && activeIcon()" class="relative flex justify-start items-center h-full">
				<div class="menu-entry-icon h-6 *:h-full" v-html="ui.icon.v"></div>
				<div class="menu-entry-icon-active h-6 *:h-full" v-html="activeIcon()"></div>
				<!-- Optional red badge -->
				<div
					v-if="ui.badge.v"
					class="absolute -top-1.5 -right-1.5 flex justify-center items-center h-5 px-1 rounded-full bg-error"
				>
					<p class="text-sm text-white">{{ ui.badge.v }}</p>
				</div>
			</div>
			<div class="flex justify-start items-center h-6">
				<p class="grow leading-tight select-none align-bottom">{{ ui.title.v }}</p>
			</div>
			<TriangleDown v-if="hasSubMenuEntries" class="shrink-0 basis-2" :class="triangleClass" />
		</div>

		<template v-if="ui.expanded.v">
			<div class="flex flex-col justify-start items-start gap-y-4 w-full pl-4">
				<BurgerMenuEntry
					v-for="(menuEntry, index) in ui.menu.v"
					:key="index" :ui="menuEntry"
					:top-level="false"
					@clicked="$emit('clicked')"
				/>
			</div>
		</template>
	</div>
</template>

<script setup lang="ts">
import TriangleDown from '@/assets/svg/triangleDown.svg';
import type { MenuEntry } from '@/shared/protocol/ora/menuEntry';
import { computed } from 'vue';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import type {SVG} from "@/shared/protocol/ora/sVG";

const props = defineProps<{
	ui: MenuEntry;
	topLevel: boolean;
}>();

const emit = defineEmits<{
	(e: 'clicked'): void;
}>();

const serviceAdapter = useServiceAdapter();

const hasSubMenuEntries = computed((): boolean => {
	return props.ui.menu.v && props.ui.menu.v.length > 0;
});

const menuEntryClickable = computed((): boolean => hasSubMenuEntries.value || !!props.ui.action.v);

const menuEntryActive = computed((): boolean => {
	return `/${props.ui.componentFactoryId.v}` === window.location.pathname;
});

const triangleClass = computed((): string|null => {
	if (props.topLevel) {
		return '-rotate-90';
	}
	if (props.ui.expanded.v) {
		return 'rotate-180';
	}
	return null;
})

function activeIcon(): SVG{
	if (props.ui.iconActive.v){
		return props.ui.iconActive.v
	}

	return props.ui.icon.v
}

function menuEntryClicked(): void {
	if (hasSubMenuEntries.value) {
		expandMenuEntry();
		return;
	}
	if (props.ui.action.v) {
		emit('clicked');
		serviceAdapter.executeFunctions(props.ui.action);
	}
}

function expandMenuEntry(): void {
	if (hasSubMenuEntries.value) {
		serviceAdapter.setPropertiesAndCallFunctions([
			{
				...props.ui.expanded,
				v: true,
			},
		], [props.ui.onFocus]);
	}
}
</script>

<style scoped>
.menu-entry:hover .menu-entry-icon,
.menu-entry:focus-visible .menu-entry-icon,
.menu-entry .menu-entry-icon-active {
	@apply hidden;
}

.menu-entry:hover .menu-entry-icon-active,
.menu-entry:focus-visible .menu-entry-icon-active {
	@apply block;
}
</style>
