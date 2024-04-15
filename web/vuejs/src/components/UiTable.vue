<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import type { LivePage } from '@/shared/model/livePage';
import {Table} from "@/shared/protocol/gen/table";

const props = defineProps<{
	ui: Table;
	page: LivePage;
}>();
</script>

<template>
	<div class="relative overflow-x-auto shadow-md">
		<table class="w-full text-left text-sm text-gray-500 rtl:text-right dark:text-gray-400">
			<thead
				v-if="props.ui.headers.v"
				class="bg-gray-50 text-xs uppercase text-gray-700 dark:bg-gray-700 dark:text-gray-400"
			>
				<tr>
					<th v-for="head in props.ui.headers.v" scope="col" class="px-6 py-3">
						<ui-generic :ui="head.body.v" :page="page" />
					</th>
				</tr>
			</thead>

			<tbody>
				<tr
					v-for="row in props.ui.rows.v"
					class="border-b odd:bg-white even:bg-gray-50 dark:border-gray-700 odd:dark:bg-gray-900 even:dark:bg-gray-800"
				>
					<td v-for="cell in row.cells.v" class="px-6 py-4">
						<ui-generic :ui="cell.body.v" :page="page" />
					</td>
				</tr>
			</tbody>
		</table>
	</div>
</template>
