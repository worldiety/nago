<script lang="ts" setup>
import {isNil} from "@/shared/protocol/util";
import {Str} from "@/shared/protocol/ora/str";

const props = defineProps<{
	ui: Str;
}>();

</script>

<template>
	<span v-if="!isNil(ui.v)">{{ props.ui.v }}</span>
</template>
