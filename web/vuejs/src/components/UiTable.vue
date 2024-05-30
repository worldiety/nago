<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {Table} from "@/shared/protocol/ora/table";

const props = defineProps<{
	ui: Table;
}>();
</script>

<template>
	<div class="relative overflow-x-auto shadow-md rounded-lg">
		<table class="w-full text-left text-sm rtl:text-right">
			<thead
				v-if="props.ui.headers.v"
				class="text-xs uppercase bg-primary-87 darkmode:bg-primary-30"
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
					class="even:bg-primary-94 odd:bg-primary-92 darkmode:even:bg-primary-14 darkmode:odd:bg-primary-17"
				>
					<td v-for="cell in row.cells.v" class="px-6 py-4">
						<ui-generic :ui="cell.body.v"  />
					</td>
				</tr>
			</tbody>
		</table>
	</div>
</template>
