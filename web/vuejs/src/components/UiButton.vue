<script lang="ts" setup>
import { computed } from 'vue';
import type { Button } from '@/shared/protocol/gen/button';
import { useServiceAdapter } from '@/composables/serviceAdapter';

const props = defineProps<{
	ui: Button;
}>();

const serviceAdapter = useServiceAdapter();

function onClick() {
	serviceAdapter.executeFunctions(props.ui.action);
}

const buttonClasses = computed<string>(() => {
	const classes: string[] = [];
	switch (props.ui.color.v) {
		case 'primary':
			classes.push('button-primary');
			break;
		case 'secondary':
			classes.push('button-secondary');
			break;
		case 'tertiary':
			classes.push('button-tertiary');
			break;
		case 'destructive':
			classes.push('button-destructive');
			break;
		default:
			classes.push('button-default');
	}
	if (iconOnly.value) {
		// Make button round when it shows an icon only
		classes.push('!p-0 !w-10');
	}
	return classes.join(' ');
});

const iconOnly = computed<boolean>(() => {
	return props.ui.caption.v == '' && props.ui.preIcon.v != '';
});
</script>

<template>
	<button :class="buttonClasses" :disabled="props.ui.disabled.v" @click="onClick">
		<svg v-if="iconOnly" v-inline class="h-4 w-4" v-html="props.ui.preIcon.v"></svg>
		<template v-else>
			<svg v-if="props.ui.preIcon.v" class="mr-2 h-4 w-4" v-html="props.ui.preIcon.v"></svg>
			<span>{{ props.ui.caption.v }}</span>
			<svg v-if="props.ui.postIcon.v" class="ml-2 h-4 w-4" v-html="props.ui.postIcon.v"></svg>
		</template>
	</button>
</template>
