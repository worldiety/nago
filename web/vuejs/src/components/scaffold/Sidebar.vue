<template>
	<nav
		ref="sidebar"
		class="fixed top-0 left-0 bottom-0 text-black dark:text-white h-full w-32 z-30"
		aria-label="Sidebar"
	>
		<!-- Sidebar -->
		<div class="relative flex flex-col items-center justify-start gap-y-4 bg-white dark:bg-darkmode-gray h-full w-full pt-6 px-4 pb-7 z-10">
			<div class="w-full *:w-full" v-html="ui.logo.v"></div>
			<!-- Top level menu entries -->
			<div class="flex flex-col gap-y-4 justify-start items-center overflow-y-auto h-full w-full">
				<div v-for="(menuEntry, index) in ui.menu.v" :key="index" ref="menuEntryElements" class="w-full">
					<MenuEntryComponent
						:ui="menuEntry"
						:menu-entry-index="index"
						:mode="'sidebar'"
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
				ref="subMenu"
				class="absolute top-0 left-32 bottom-0 flex flex-col justify-start gap-y-4 bg-white dark:bg-darkmode-gray border-l border-l-disabled-background dark:border-l-disabled-text rounded-r-2xl shadow-md w-72 py-8 px-2 z-0"
			>
				<!-- Sub menu entries -->
				<div
					v-for="(subMenuEntry, subMenuEntryIndex) in subMenuEntries"
					:key="subMenuEntryIndex"
					class="flex flex-col justify-start gap-y-2"
				>
					<div
						ref="subMenuEntryElements"
						class="flex justify-between items-center hover:bg-disabled-background hover:bg-opacity-25 active:bg-opacity-35 rounded-full py-2 px-4"
						:class="{'cursor-pointer': isClickableMenuEntry(subMenuEntry)}"
						:tabindex="isClickableMenuEntry(subMenuEntry) ? '0' : '-1'"
						@click="menuEntryClicked(subMenuEntry)"
						@keydown.enter="menuEntryClicked(subMenuEntry)"
					>
						<p class="font-medium">{{ subMenuEntry.title.v }}</p>
						<TriangleDown
							v-if="subMenuEntry.menu.v?.length > 0"
							class="duration-150 w-2 -mr-1"
							:class="{'rotate-180': subMenuEntry.expanded.v}"
						/>
					</div>
					<div
						v-if="subMenuEntry.expanded.v && subMenuEntry.menu.v?.length > 0"
						class="flex flex-col justify-start gap-y-2 pl-4"
					>
						<!-- Sub sub menu entries -->
						<p
							v-for="(subSubMenuEntry, subSubMenuEntryIndex) in getSubSubMenuEntries(subMenuEntry)"
							:key="subSubMenuEntryIndex"
							ref="subSubMenuEntryElements"
							class="hover:bg-disabled-background hover:bg-opacity-25 active:bg-opacity-35 rounded-full py-2 px-4"
							:class="{'cursor-pointer': subSubMenuEntry.action.v}"
							:tabindex="subSubMenuEntry.action.v ? '0' : '-1'"
							@click="menuEntryClicked(subSubMenuEntry)"
							@keydown.enter="menuEntryClicked(subSubMenuEntry)"
						>
							{{ subSubMenuEntry.title.v }}
						</p>
					</div>
				</div>
			</div>
		</Transition>
	</nav>
</template>

<script setup lang="ts">
import type { NavigationComponent } from '@/shared/protocol/ora/navigationComponent';
import ThemeToggle from '@/components/scaffold/ThemeToggle.vue';
import MenuEntryComponent from '@/components/scaffold/TopLevelMenuEntry.vue';
import { computed, onMounted, onUnmounted, ref } from 'vue';
import type { MenuEntry } from '@/shared/protocol/ora/menuEntry';
import TriangleDown from '@/assets/svg/triangleDown.svg';
import { useServiceAdapter } from '@/composables/serviceAdapter';

const props = defineProps<{
	ui: NavigationComponent;
}>();

const serviceAdapter = useServiceAdapter();
const sidebar = ref<HTMLElement|undefined>();
const subMenu = ref<HTMLElement|undefined>();
const subMenuEntryElements = ref<HTMLElement[]>([]);
const subSubMenuEntryElements = ref<HTMLElement[]>([]);

onMounted(() => {
	document.addEventListener('mousemove', handleMouseMove);
});

onUnmounted(() => {
	document.removeEventListener('mousemove', handleMouseMove);
});

const expandedMenuEntry = computed((): MenuEntry|undefined => {
	return props.ui.menu.v?.find((menuEntry) => menuEntry.expanded.v);
});

const subMenuEntries = computed((): MenuEntry[] => {
	const entries: MenuEntry[] = props.ui.menu.v
		?.filter((menuEntry) => menuEntry.expanded.v)
		.flatMap((menuEntry) => menuEntry.menu.v ?? []);
	// Add the expanded menu entry without its sub menu entries, if it has an action
	if (entries.length > 0 && expandedMenuEntry.value?.action.v) {
		entries.unshift({
			...expandedMenuEntry.value,
			menu: {
				...expandedMenuEntry.value.menu,
				v: [],
			}
		});
	}
	return entries;
});

function isClickableMenuEntry(menuEntry: MenuEntry): boolean {
	// Clickable, if it has an action or sub menu entries
	return !!menuEntry.action.v || menuEntry.menu.v && menuEntry.menu.v.length > 0;
}

function isLinkingMenuEntry(menuEntry: MenuEntry): boolean {
	// Linking, if it has an action and no sub menu entries
	return !!menuEntry.action.v && (!menuEntry.menu.v || menuEntry.menu.v.length === 0);
}

function handleMouseMove(event: MouseEvent): void {
	const threshold = subMenu.value?.getBoundingClientRect().right
		?? sidebar.value?.getBoundingClientRect().right
		?? 0;
	if (event.x > threshold) {
		// Collapse the sub menu when threshold is passed
		const updatedExpandedProperties = props.ui.menu.v
			?.filter((menuEntry) => menuEntry.expanded.v)
			.map((menuEntry) => ({
				...menuEntry.expanded,
				v: false,
			}));
		if (updatedExpandedProperties.length > 0) {
			serviceAdapter.setProperties(...updatedExpandedProperties);
		}
	}
}

function focusFirstLinkedSubMenuEntry(): void {
	const elementToFocus =
		subMenuEntryElements.value.find((subMenuEntryElement) => subMenuEntryElement.tabIndex === 0)
		?? subSubMenuEntryElements.value.find((subMenuEntryElement) => subMenuEntryElement.tabIndex === 0);
	elementToFocus?.focus();
}

function menuEntryClicked(menuEntry: MenuEntry): void {
	if (isClickableMenuEntry(menuEntry)) {
		if (menuEntry.menu.v && menuEntry.menu.v.length > 0) {
			serviceAdapter.setProperties({
				...menuEntry.expanded,
				v: !menuEntry.expanded.v,
			});
		} else if (menuEntry.action.v) {
			serviceAdapter.executeFunctions(menuEntry.action);
		}
	}
}

function getSubSubMenuEntries(subMenuEntry: MenuEntry): MenuEntry[] {
	const entries: MenuEntry[] = [...subMenuEntry.menu.v];
	// Add the sub menu entry without its sub menu entries, if it has an action
	if (entries.length > 0 && subMenuEntry.action.v) {
		entries.unshift({
			...subMenuEntry,
			menu: {
				...subMenuEntry.menu,
				v: [],
			}
		});
	}
	return entries;
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
