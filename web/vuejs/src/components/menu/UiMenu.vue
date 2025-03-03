<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {useServiceAdapter} from '@/composables/serviceAdapter';
import {nextRID} from '@/eventhandling';
import {Menu, MenuButton, MenuItem, MenuItems} from '@headlessui/vue';
import {FunctionCallRequested, Menu as ProtoMenu, MenuItem as ProtoMenuItem} from '@/shared/proto/nprotoc_gen';
import {usePopper} from "@/shared/use-popper";


const props = defineProps<{
	ui: ProtoMenu;
}>();

const serviceAdapter = useServiceAdapter();

let [trigger, container] = usePopper({
	placement: "bottom-end",
	strategy: "fixed",
	modifiers: [
		{name: "flip", enabled: true},
		{name: "offset", options: {offset: [0, 8]}},


	],
});

function itemClick(item: ProtoMenuItem) {
	if (item.action) {
		serviceAdapter.sendEvent(new FunctionCallRequested(item.action, nextRID()));
		return;
	}
}
</script>

<template>
	<Menu as="div" class="relative inline-block text-left" >
		<div>
			<MenuButton class="inline-flex w-full justify-center" ref="trigger">
				<ui-generic v-if="props.ui.anchor" :ui="props.ui.anchor"/>
			</MenuButton>
		</div>


			<transition ref="container" style="z-index: 40"
		
			>
				<MenuItems
					class="absolute right-0 mt-2 w-56 origin-top-right divide-y divide-M3 rounded-md bg-M1 shadow-lg ring-1 ring-black/5 focus:outline-none"
				>
					<div class="px-1 py-1" v-for="section in props.ui.groups?.value">
						<MenuItem v-for="ui in section.items?.value" v-slot="{ active }">
							<button
								:class="[
								active ? 'bg-I0 bg-opacity-25 text-M8' : 'text-gray-900',
								'group flex w-full items-center rounded-md px-2 py-2 text-sm',
							]"
								@click="itemClick(ui)"
							>
								<ui-generic v-if="ui.content" :ui="ui.content"/>
							</button>
						</MenuItem>
					</div>
				</MenuItems>
			</transition>

	</Menu>
</template>
