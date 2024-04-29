<template>
	<NavigationBar v-if="props.ui.navigationComponent.v.alignment.v === Alignment.TOP" :ui="props.ui.navigationComponent.v" />
	<Sidebar v-if="props.ui.navigationComponent.v.alignment.v === Alignment.LEFT" :ui="props.ui.navigationComponent.v" />

	<div class="p-4">
		<ui-generic :ui="props.ui.body.v"  />
	</div>
</template>

<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import type { Scaffold } from "@/shared/protocol/gen/scaffold";
import Sidebar from '@/components/scaffold/Sidebar.vue';
import NavigationBar from '@/components/scaffold/NavigationBar.vue';
import { Alignment } from '@/shared/protocol/alignment';

const props = defineProps<{
	ui: Scaffold;
}>();
</script>
