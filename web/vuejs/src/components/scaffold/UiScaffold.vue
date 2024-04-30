<template>
	<NavigationBar v-if="navigationBarVisible" :ui="props.ui.navigationComponent.v" />
	<Sidebar v-if="sidebarVisible" :ui="props.ui.navigationComponent.v" />

	<div class="website-content" :class="bodyWrapperClass">
		<ui-generic :ui="props.ui.body.v"  />
	</div>
</template>

<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import type { Scaffold } from "@/shared/protocol/gen/scaffold";
import Sidebar from '@/components/scaffold/Sidebar.vue';
import NavigationBar from '@/components/scaffold/NavigationBar.vue';
import { Alignment } from '@/shared/protocol/alignment';
import { computed } from 'vue';

const props = defineProps<{
	ui: Scaffold;
}>();

const navigationBarVisible = computed((): boolean => props.ui.navigationComponent.v.alignment.v === Alignment.TOP);

const sidebarVisible = computed((): boolean => props.ui.navigationComponent.v.alignment.v === Alignment.LEFT);

const bodyWrapperClass = computed((): string => {
	if (navigationBarVisible.value) {
		return 'pt-16';
	}
});
</script>
