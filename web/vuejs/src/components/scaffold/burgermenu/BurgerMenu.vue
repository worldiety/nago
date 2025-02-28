<template>
	<div class="fixed top-0 left-0 right-0 text-black bg-M1 border-b border-b-M5 h-24 py-4 px-8 z-30">
		<!-- Top bar -->
		<div class="relative flex justify-start items-center h-full">
			<MenuIcon
				class="relative cursor-pointer h-6 z-10"
				tabindex="0"
				@click="menuOpen = true"
				@keydown.enter="menuOpen = true"
			/>
			<div class="absolute top-0 left-0 bottom-0 right-0 flex justify-center items-center h-full z-0">
				<div class="">
					<ui-generic v-if="props.ui.logo" :ui="props.ui.logo" />
				</div>
			</div>
		</div>

		<!-- Menu -->
		<Transition name="slide">
			<div
				v-if="menuOpen"
				class="fixed top-0 left-0 bottom-0 flex flex-col justify-start items-start w-full xs:w-80 bg-M4 shadow-md z-20"
			>
				<div class="flex justify-start items-center h-24 p-8">
					<CloseIcon
						tabindex="0"
						class="cursor-pointer h-6"
						@click="menuOpen = false"
						@keydown.enter="menuOpen = false"
					/>
				</div>
				<div class="flex flex-col justify-start items-start gap-y-4 overflow-y-auto basis-full w-full p-4">
					<template v-if="!subMenuVisible">
						<!-- Top level menu entries -->
						<BurgerMenuEntry
							v-for="(menuEntry, index) in ui.menu?.value"
							:key="index"
							:ui="menuEntry"
							:top-level="true"
							@clicked="menuOpen = false"
						/>
					</template>
					<div v-else class="flex flex-col justify-start items-start gap-y-4 w-full pl-4">
						<!-- Back to top level menu button -->
						<div
							tabindex="0"
							class="relative flex justify-start items-center gap-x-2 cursor-pointer hover:bg-disabled-background hover:bg-opacity-25 active:bg-opacity-35 rounded-full w-full py-4 pl-8 pr-4 -ml-4"
							@click="returnToTopLevelMenu"
							@keydown.enter="returnToTopLevelMenu"
						>
							<div class="h-6">
								<TriangleDown class="absolute top-0 left-4 bottom-0 rotate-90 h-2 my-auto" />
							</div>
							<p class="leading-tight font-semibold">{{ $t('scaffold.toMenu') }}</p>
						</div>
						<!-- Top level menu entry title button -->

						<div
							:tabindex="expandedTopLevelMenuEntryLinked ? '0' : '-1'"
							class="flex justify-start items-center gap-x-2 rounded-full w-full p-4"
							:class="{
								'cursor-pointer hover:bg-disabled-background hover:bg-opacity-25 active:bg-opacity-35':
									expandedTopLevelMenuEntryLinked,
								'bg-disabled-background bg-opacity-35': expandedTopLevelMenuEntryActive,
							}"
							@click="navigateToExpandedTopLevelMenuEntry"
							@keydown.enter="navigateToExpandedTopLevelMenuEntry"
						>
							<div class="flex justify-start items-center h-6">
								<p class="leading-tight font-semibold">{{ expandedTopLevelMenuEntry?.title }}</p>
							</div>
						</div>
						<!-- Sub menu entries -->
						<BurgerMenuEntry
							v-for="(menuEntry, index) in subMenuEntries"
							:key="index"
							:ui="menuEntry"
							:top-level="false"
							@clicked="menuOpen = false"
						/>
					</div>
				</div>
				<div class="flex justify-center items-center w-full p-4">
					<ThemeToggle />
				</div>
			</div>
		</Transition>
	</div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import CloseIcon from '@/assets/svg/closeBold.svg';
import MenuIcon from '@/assets/svg/menu.svg';
import TriangleDown from '@/assets/svg/triangleDown.svg';
import UiGeneric from '@/components/UiGeneric.vue';
import ThemeToggle from '@/components/scaffold/ThemeToggle.vue';
import BurgerMenuEntry from '@/components/scaffold/burgermenu/BurgerMenuEntry.vue';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { FunctionCallRequested, Scaffold, ScaffoldMenuEntry } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Scaffold;
}>();

const serviceAdapter = useServiceAdapter();
const menuOpen = ref<boolean>(false);

const expandedTopLevelMenuEntry = computed((): ScaffoldMenuEntry | null => {
	if (!props.ui.menu) {
		return null;
	}
	return props.ui.menu.value?.find((menuEntry: ScaffoldMenuEntry) => menuEntry.expanded) ?? null;
});

const expandedTopLevelMenuEntryLinked = computed((): boolean => {
	return !!expandedTopLevelMenuEntry.value && expandedTopLevelMenuEntry.value.action == undefined;
});

const expandedTopLevelMenuEntryActive = computed((): boolean => {
	return (
		!!expandedTopLevelMenuEntry.value && `/${expandedTopLevelMenuEntry.value.rootView}` === window.location.pathname
	);
});

const subMenuVisible = computed((): boolean => {
	if (!props.ui.menu) {
		return false;
	}

	const expandedTopLevelMenuEntry = props.ui.menu.value?.find((menuEntry: ScaffoldMenuEntry) => menuEntry.expanded);

	if (!expandedTopLevelMenuEntry?.menu) {
		return false;
	}

	return !!expandedTopLevelMenuEntry?.menu.value;
});

const subMenuEntries = computed((): ScaffoldMenuEntry[] => {
	if (!props.ui.menu) {
		return [];
	}
	const expandedTopLevelMenuEntry = props.ui.menu.value?.find((menuEntry: ScaffoldMenuEntry) => menuEntry.expanded);

	if (!expandedTopLevelMenuEntry) {
		return props.ui.menu.value;
	}
	return expandedTopLevelMenuEntry.menu?.value ?? props.ui.menu.value;
});

function navigateToExpandedTopLevelMenuEntry(): void {
	if (!expandedTopLevelMenuEntry.value?.action == undefined) {
		return;
	}

	if (expandedTopLevelMenuEntry.value) {
		menuEntryClicked(expandedTopLevelMenuEntry.value);
	}
}

function menuEntryClicked(menuEntry: ScaffoldMenuEntry): void {
	if (!menuEntry.action) {
		return;
	}

	serviceAdapter.sendEvent(new FunctionCallRequested(menuEntry.action, nextRID()));
}

function returnToTopLevelMenu(): void {
	if (!expandedTopLevelMenuEntry.value) {
		return;
	}
	// serviceAdapter.setPropertiesAndCallFunctions([{
	// 	...expandedTopLevelMenuEntry.value.x,
	// 	v: false,
	// }], [expandedTopLevelMenuEntry.value.onFocus])

	expandedTopLevelMenuEntry.value.expanded = false;
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
