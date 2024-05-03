<template>
	<nav
		class="fixed top-0 left-0 bottom-0 text-black dark:text-white h-full w-32 z-10"
		aria-label="Sidebar"
	>
		<!-- Sidebar -->
		<div class="relative flex flex-col justify-between items-center bg-white dark:bg-darkmode-gray h-full py-8 px-2 z-10">
			<div class="flex flex-col items-center justify-start gap-y-4">
				<div class="*:w-full mb-4" v-html="ui.logo.v"></div>
				<div v-for="(menuEntry, index) in ui.menu.v" :key="index" ref="menuEntryElements" :data-index="index">
					<MenuEntryComponent
						:ui="menuEntry"
						:menu-entry-index="index"
						:expanded="menuEntry.id === activeMenuEntry?.id"
						@expand-menu-entry="expandMenuEntry"
						@collapse-menu-entry="collapseMenuEntry"
						@focus-first-linked-sub-menu-entry="focusFirstLinkedSubMenuEntry"
					/>
				</div>
			</div>
			<ThemeToggle />
		</div>

		<!-- Sub menu -->
		<Transition name="slide">
			<div
				v-if="subMenuEntries.length > 0"
				class="absolute top-0 left-32 bottom-0 bg-white dark:bg-darkmode-gray border-l border-l-disabled-background dark:border-l-disabled-text rounded-r-2xl shadow-md w-72 py-8 px-4 z-0"
			>
				<div v-for="(subMenuEntry, subMenuEntryIndex) in subMenuEntries" :key="subMenuEntryIndex">
					<p
						ref="subMenuEntryElements"
						class="font-medium"
						:class="{
								'mb-4': subMenuEntry.menu.v?.length > 0,
								'cursor-pointer hover:underline focus-visible:underline': subMenuEntry.action.v,
							}"
						:tabindex="subMenuEntry.action.v ? '0' : '-1'"
						@click="menuEntryClicked(subMenuEntry.action)"
						@keydown.enter="menuEntryClicked(subMenuEntry.action)"
					>
						{{ subMenuEntry.title.v }}
					</p>
					<p
						v-for="(subSubMenuEntry, subSubMenuEntryIndex) in subMenuEntry.menu.v"
						:key="subSubMenuEntryIndex"
						:class="{'cursor-pointer hover:underline focus-visible:underline': subSubMenuEntry.action.v}"
						:tabindex="subSubMenuEntry.action.v ? '0' : '-1'"
						@click="menuEntryClicked(subSubMenuEntry.action)"
						@keydown.enter="menuEntryClicked(subSubMenuEntry.action)"
					>
						{{ subSubMenuEntry.title.v }}
					</p>
				</div>
			</div>
		</Transition>
	</nav>
</template>

<script setup lang="ts">
import type { NavigationComponent } from '@/shared/protocol/gen/navigationComponent';
import ThemeToggle from '@/components/scaffold/ThemeToggle.vue';
import MenuEntryComponent from '@/components/scaffold/MenuEntryComponent.vue';
import { computed, ref } from 'vue';
import type { MenuEntry } from '@/shared/protocol/gen/menuEntry';
import type { Property } from '@/shared/protocol/property';
import type { Pointer } from '@/shared/protocol/pointer';
import { useServiceAdapter } from '@/composables/serviceAdapter';

defineProps<{
	ui: NavigationComponent;
}>();

const serviceAdapter = useServiceAdapter();
const subMenuEntryElements = ref<HTMLElement[]>([]);
const activeMenuEntry = ref<MenuEntry|null>(null);
const activeMenuEntryIndex = ref<number|null>(null);

const subMenuEntries = computed((): MenuEntry[] => activeMenuEntry.value?.menu.v ?? []);

function expandMenuEntry(menuEntry: MenuEntry, menuEntryIndex: number): void {
	setActiveMenuEntry(menuEntry, menuEntryIndex);
}

function collapseMenuEntry(): void {
	setActiveMenuEntry(null, null);
}

function setActiveMenuEntry(menuEntry: MenuEntry|null, menuEntryIndex: number|null): void {
	activeMenuEntry.value = menuEntry;
	activeMenuEntryIndex.value = menuEntryIndex;
}

function focusFirstLinkedSubMenuEntry(): void {
	subMenuEntryElements.value.find((subMenuEntryElement) => subMenuEntryElement.tabIndex === 0)?.focus();
}

function menuEntryClicked(menuEntryAction: Property<Pointer>): void {
	if (menuEntryAction.v) {
		serviceAdapter.executeFunctions(menuEntryAction);
	}
}
</script>

<style scoped>
/* Vue transitions: https://vuejs.org/guide/built-ins/transition#css-transitions */
.slide-enter-active,
.slide-leave-active {
	@apply transform duration-200 ease-in-out;
}

.slide-enter-from,
.slide-leave-to {
	@apply translate-x-[-100%];
}
</style>
