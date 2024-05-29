<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {Table} from "@/shared/protocol/ora/table";

const props = defineProps<{
	ui: Table;
}>();
</script>

<template>
	<div class="relative overflow-x-auto shadow-md">
		<table class="w-full text-left text-sm text-gray-500 rtl:text-right">
			<thead
				v-if="props.ui.headers.v"
				class="bg-gray-50 text-xs uppercase text-gray-700"
			>
				<tr>
					<th v-for="head in props.ui.headers.v" scope="col" class="px-6 py-3">
						<ui-generic :ui="head.body.v" />
					</th>
				</tr>
			</thead>

			<tbody>
				<tr
					v-for="row in props.ui.rows.v"
					class="border-b odd:bg-white even:bg-gray-50"
				>
					<td v-for="cell in row.cells.v" class="px-6 py-4">
						<ui-generic :ui="cell.body.v"  />
					</td>
				</tr>
			</tbody>
		</table>
	</div>
</template>
