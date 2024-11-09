<template>
	<div class="flex flex-col justify-start items-start gap-y-4 w-full">
		<div
			class="menu-entry flex justify-start items-center gap-x-4 rounded-full w-full p-4"
			:class="{
				'cursor-pointer hover:bg-M7 hover:bg-opacity-25 active:bg-opacity-35': menuEntryClickable,
				'bg-M7 bg-opacity-35': menuEntryActive,
			}"
			:tabindex="menuEntryClickable ? '0' : '-1'"
			@click="menuEntryClicked"
			@keydown.enter="menuEntryClicked"
		>
			<div v-if="props.ui.i" class="relative flex justify-start items-center h-full">

				<div class="menu-entry-icon-active *:h-full">
					<ui-generic v-if="ui.x && props.ui.v" :ui="props.ui.v"/>
					<ui-generic v-else :ui="props.ui.i"/>
				</div>
				<div class="menu-entry-icon  *:h-full">
					<ui-generic :ui="props.ui.i"/>
				</div>


				<!-- Optional red badge -->
				<div
					v-if="ui.b"
					class="absolute -top-1.5 -right-1.5 flex justify-center items-center h-5 px-1 rounded-full bg-A1"
				>
					<p class="text-sm text-white">{{ ui.b }}</p>
				</div>
			</div>
			<div class="flex justify-start items-center h-6">
				<p class="grow leading-tight select-none align-bottom">{{ ui.t }}aaa</p>
			</div>
			<TriangleDown v-if="hasSubMenuEntries" class="shrink-0 basis-2" :class="triangleClass"/>
		</div>
		<template v-if="ui.x">
			<div class="flex flex-col justify-start items-start gap-y-4 w-full pl-4">
				<BurgerMenuEntry
					v-for="(menuEntry, index) in ui.m"
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
import {computed} from 'vue';
import {useServiceAdapter} from '@/composables/serviceAdapter';
import {ScaffoldMenuEntry} from "@/shared/protocol/ora/scaffoldMenuEntry";
import UiGeneric from "@/components/UiGeneric.vue";

const props = defineProps<{
	ui: ScaffoldMenuEntry;
	topLevel: boolean;
}>();

const emit = defineEmits<{
	(e: 'clicked'): void;
}>();

const serviceAdapter = useServiceAdapter();

const hasSubMenuEntries = computed((): boolean => {
	return !!(props.ui.m && props.ui.m.length > 0);
});

const menuEntryClickable = computed((): boolean => hasSubMenuEntries.value || !!props.ui.a);

const menuEntryActive = computed((): boolean => {
	//return true
	if (props.ui.f == "." && (window.location.pathname == "" || window.location.pathname == "/")) {
		return true
	}

	return `/${props.ui.f}` === window.location.pathname;
});

const triangleClass = computed((): string | null => {
	if (props.topLevel) {
		return '-rotate-90';
	}
	if (props.ui.x) {
		return 'rotate-180';
	}
	return null;
})


function menuEntryClicked(): void {
	if (hasSubMenuEntries.value) {
		expandMenuEntry();
		return;
	}
	if (props.ui.a) {
		emit('clicked');
		serviceAdapter.executeFunctions(props.ui.a);
	}
}

function expandMenuEntry(): void {
	// if (hasSubMenuEntries.value) {
	// 	serviceAdapter.setPropertiesAndCallFunctions([
	// 		{
	// 			...props.ui.x,
	// 			v: true,
	// 		},
	// 	], [props.ui.onFocus]);
	// }

	props.ui.x = true
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
