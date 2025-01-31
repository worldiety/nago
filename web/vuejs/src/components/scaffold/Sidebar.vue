<template>
	<nav ref="sidebar" class="fixed top-0 left-0 bottom-0 text-black h-full w-32 z-30 bg-M4" aria-label="Sidebar">
		<!-- Sidebar -->
		<div class="relative flex flex-col items-center justify-start gap-y-4 h-full w-full pt-6 px-4 pb-7 z-10 bg-M4">
			<div v-if="ui.l" class="w-full *:w-full mb-4">
				<ui-generic :ui="ui.l" />
			</div>
			<!-- Top level menu entries -->
			<div class="flex flex-col gap-y-4 justify-start items-center overflow-y-auto h-full w-full">
				<div v-for="(menuEntry, index) in ui.m" :key="index" ref="menuEntryElements" class="w-full">
					<MenuEntryComponent
						:ui="menuEntry"
						:menu-entry-index="index"
						:mode="'sidebar'"
						@focus-first-linked-sub-menu-entry="focusFirstLinkedSubMenuEntry"
						@expand="expandMenuEntry"
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
				class="absolute top-0 left-32 bottom-0 flex flex-col justify-start gap-y-4 bg-M1 border-l border-l-M5 rounded-r-2xl shadow-md w-72 py-8 px-2 z-0 bg-M4"
			>
				<!-- Sub menu entries -->
				<div
					v-for="(subMenuEntry, subMenuEntryIndex) in subMenuEntries"
					:key="subMenuEntryIndex"
					class="flex flex-col justify-start gap-y-2"
				>
					<div
						ref="subMenuEntryElements"
						class="flex justify-between items-center rounded-full py-2 px-4"
						:class="{
							'cursor-pointer hover:bg-disabled-background hover:bg-opacity-25 active:bg-opacity-35':
								isClickableMenuEntry(subMenuEntry),
							'bg-disabled-background bg-opacity-35': isActiveMenuEntry(subMenuEntry),
						}"
						:tabindex="isClickableMenuEntry(subMenuEntry) ? '0' : '-1'"
						@click="menuEntryClicked(subMenuEntry)"
						@keydown.enter="menuEntryClicked(subMenuEntry)"
					>
						<p class="font-medium">{{ subMenuEntry.t }}</p>
						<TriangleDown
							v-if="subMenuEntry.m?.length > 0"
							class="duration-150 w-2 -mr-1"
							:class="{ 'rotate-180': subMenuEntry.x }"
						/>
					</div>
					<div
						v-if="subMenuEntry.x && subMenuEntry.m?.length > 0"
						class="flex flex-col justify-start gap-y-2 pl-4"
					>
						<!-- Sub sub menu entries -->
						<p
							v-for="(subSubMenuEntry, subSubMenuEntryIndex) in getSubSubMenuEntries(subMenuEntry)"
							:key="subSubMenuEntryIndex"
							ref="subSubMenuEntryElements"
							class="rounded-full py-2 px-4"
							:class="{
								'cursor-pointer hover:bg-disabled-background hover:bg-opacity-25 active:bg-opacity-35':
									subSubMenuEntry.a,
								'bg-disabled-background bg-opacity-35': isActiveMenuEntry(subSubMenuEntry),
							}"
							:tabindex="subSubMenuEntry.a ? '0' : '-1'"
							@click="menuEntryClicked(subSubMenuEntry)"
							@keydown.enter="menuEntryClicked(subSubMenuEntry)"
						>
							{{ subSubMenuEntry.t }}
						</p>
					</div>
				</div>
			</div>
		</Transition>
	</nav>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import TriangleDown from '@/assets/svg/triangleDown.svg';
import UiGeneric from '@/components/UiGeneric.vue';
import ThemeToggle from '@/components/scaffold/ThemeToggle.vue';
import MenuEntryComponent from '@/components/scaffold/TopLevelMenuEntry.vue';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { Scaffold } from '@/shared/protocol/ora/scaffold';
import { ScaffoldMenuEntry } from '@/shared/protocol/ora/scaffoldMenuEntry';

const props = defineProps<{
	ui: Scaffold;
}>();

const serviceAdapter = useServiceAdapter();
const sidebar = ref<HTMLElement | undefined>();
const subMenu = ref<HTMLElement | undefined>();
const subMenuEntryElements = ref<HTMLElement[]>([]);
const subSubMenuEntryElements = ref<HTMLElement[]>([]);

onMounted(() => {
	document.addEventListener('mousemove', handleMouseMove);
});

onUnmounted(() => {
	document.removeEventListener('mousemove', handleMouseMove);
});

const expandedMenuEntry = computed((): ScaffoldMenuEntry | undefined => {
	return props.ui.m?.find((menuEntry) => menuEntry.x);
});

const subMenuEntries = computed((): ScaffoldMenuEntry[] => {
	const entries: ScaffoldMenuEntry[] = props.ui.m
		?.filter((menuEntry) => menuEntry.x)
		.flatMap((menuEntry) => menuEntry.m ?? []);
	// Add the expanded menu entry without its sub menu entries, if it has an action
	if (entries?.length > 0 && expandedMenuEntry.value?.a) {
		entries.unshift({
			...expandedMenuEntry.value,
			m: {
				...expandedMenuEntry.value.m,
				//v: [],
			},
		});
	}
	return entries ?? [];
});

function isClickableMenuEntry(menuEntry: ScaffoldMenuEntry): boolean {
	// Clickable, if it has an action or sub menu entries
	let clickable = menuEntry.a != undefined || (menuEntry.m && menuEntry.m.length > 0);
	return !!clickable;
}

function isActiveMenuEntry(menuEntry: ScaffoldMenuEntry): boolean {
	// Active, if its component factory ID matches the current page's path name
	return `/${menuEntry.f}` === window.location.pathname;
}

function handleMouseMove(event: MouseEvent): void {
	const threshold = subMenu.value?.getBoundingClientRect().right ?? sidebar.value?.getBoundingClientRect().right ?? 0;
	if (event.x > threshold) {
		// Collapse the sub menu when threshold is passed
		// const updatedExpandedProperties = props.ui.m
		// 	?.filter((menuEntry) => menuEntry.x)
		// 	.map((menuEntry) => ({
		// 		...menuEntry.x,
		// 		v: false,
		// 	}));
		// if (updatedExpandedProperties?.length > 0) {
		// 	serviceAdapter.setProperties(...updatedExpandedProperties);
		// }

		props.ui.m?.forEach((value) => {
			value.x = false;
		});
	}
}

function focusFirstLinkedSubMenuEntry(): void {
	const elementToFocus =
		subMenuEntryElements.value.find((subMenuEntryElement) => subMenuEntryElement.tabIndex === 0) ??
		subSubMenuEntryElements.value.find((subMenuEntryElement) => subMenuEntryElement.tabIndex === 0);
	elementToFocus?.focus();
}

function menuEntryClicked(menuEntry: ScaffoldMenuEntry): void {
	if (isClickableMenuEntry(menuEntry)) {
		if (menuEntry.m && menuEntry.m.length > 0) {
			serviceAdapter.setProperties({
				...menuEntry.x,
				v: !menuEntry.x,
			});
		} else if (menuEntry.a) {
			serviceAdapter.executeFunctions(menuEntry.a);
		}
	}
}

function getSubSubMenuEntries(subMenuEntry: ScaffoldMenuEntry): ScaffoldMenuEntry[] {
	const entries: ScaffoldMenuEntry[] = [...subMenuEntry.m];
	// Add the sub menu entry without its sub menu entries, if it has an action
	if (entries.length > 0 && subMenuEntry.a) {
		entries.unshift({
			...subMenuEntry,
			m: {
				...subMenuEntry.m,
				v: [],
			},
		});
	}
	return entries;
}

function expandMenuEntry(menuEntry: ScaffoldMenuEntry): void {
	/*const propertiesToSet: Property<boolean>[] = props.ui.m.map((entry) => {
		return {
			...entry.x,
			v: entry.id === menuEntry.id,
		};
	});*/
	if (!props.ui.m) {
		return;
	}

	for (let i = 0; i < props.ui.m.length; i++) {
		let m = props.ui.m.at(i);
		m.x = false;
	}

	menuEntry.x = true;

	//serviceAdapter.setPropertiesAndCallFunctions(propertiesToSet, [menuEntry.onFocus]); //TODO?
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
