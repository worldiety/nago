<script lang="ts" setup>
import {computed} from 'vue';
import type {Text} from "@/shared/protocol/ora/text";
import {useServiceAdapter} from '@/composables/serviceAdapter';
import {isNil} from "@/shared/protocol/util";
import {namedColorClasses, namedColorStyles} from "@/components/shared/namedcolors";

const props = defineProps<{
	ui: Text;
}>();

const serviceAdapter = useServiceAdapter();

const clazz = computed<string>(() => {
	let classes = namedColorClasses(props.ui.color)

	if (props.ui.backgroundColor != undefined && props.ui.backgroundColor !== "") {
		classes += namedColorClasses(props.ui.backgroundColor)
	}

	return classes

});

const styles = computed<string>(() => {
	let s = namedColorStyles("color",props.ui.color)
	let c = namedColorStyles("background-color",props.ui.backgroundColor)

	return [s,c].join(";")
});

function onClick() {
	if (!isNil(props.ui.onClick.v)) {
		serviceAdapter.executeFunctions(props.ui.onClick);
	}
}

function onMouseEnter() {
	if (!isNil(props.ui.onHoverEnd.v)) {
		serviceAdapter.executeFunctions(props.ui.onHoverEnd);
	}
}

function onMouseLeave() {
	if (!isNil(props.ui.onHoverEnd.v)) {
		serviceAdapter.executeFunctions(props.ui.onHoverEnd);
	}
}
</script>

<template>
	<span v-if="ui.visible.v" :style="styles" :class="clazz" @click="onClick" @mouseenter="onMouseEnter"
				@mouseleave="onMouseLeave">{{
			props.ui.value.v
		}}</span>
</template>
