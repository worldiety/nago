<script lang="ts" setup>
import { computed } from 'vue';
import { useNetworkStore } from '@/stores/networkStore';
import type { LiveButton } from '@/shared/model/liveButton';
import type { LivePage } from '@/shared/model/livePage';

const props = defineProps<{
	ui: LiveButton;
	page: LivePage;
}>();

const networkStore = useNetworkStore();

function onClick() {
	networkStore.invokeFunc(props.ui.action);
}

const buttonClasses = computed<string>(() => {
	const classes: string[] = [];
	switch (props.ui.color.value) {
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
	return props.ui.caption.value == '' && props.ui.preIcon.value != '';
});
</script>

<template>
	<button :class="buttonClasses" :disabled="props.ui.disabled.value" @click="onClick">
		<svg v-if="iconOnly" v-inline class="h-4 w-4" v-html="props.ui.preIcon.value"></svg>
		<template v-else>
			<svg v-if="props.ui.preIcon.value" class="mr-2 h-4 w-4" v-html="props.ui.preIcon.value"></svg>
			<span>{{ props.ui.caption.value }}</span>
			<svg v-if="props.ui.postIcon.value" class="ml-2 h-4 w-4" v-html="props.ui.postIcon.value"></svg>
		</template>
	</button>
</template>
