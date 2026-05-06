<template>
	<div
		ref="page"
		class="page"
		:class="pageClasses"
		:style="`transition-duration: ${transitionDuration / 2}ms;`"
		:inert="!isActive"
		:aria-hidden="!isActive"
	>
		<div class="page-content">
			<div class="page-content-inner">
				<slot />
			</div>
		</div>
		<div v-if="img && !imageError" class="image-container" :style="`width: ${imageWidth};`">
			<img class="image" :src="img" alt="" @load="onImageLoad" @error="onImageError" />
		</div>
	</div>
	<div ref="dummy" class="page dummy" :class="pageClasses" aria-hidden="true" inert>
		<div class="page-content">
			<div class="page-content-inner">
				<slot />
			</div>
		</div>
		<div v-if="img && !imageError" class="image-container">
			<img class="image" :src="img" alt="" />
		</div>
	</div>
</template>
<script lang="ts" setup>
import { computed, ref } from 'vue';
import { ObjectFitValues } from '@/shared/proto/nprotoc_gen';

interface Props {
	pageId: string;
	activeId?: string;
	transitionDuration: number;
	img?: string;
	imgObjectFit?: number;
	vertical?: boolean;
}

interface Emits {
	(e: 'imageLoaded', url: string): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();
defineExpose({ calcPageHeight });

const page = ref<HTMLDivElement>();
const dummy = ref<HTMLDivElement>();
const imageError = ref(false);

const fullImage = computed<boolean>(
	() =>
		!props.imgObjectFit ||
		props.imgObjectFit === ObjectFitValues.None ||
		props.imgObjectFit === ObjectFitValues.Auto
);
const imageWidth = ref(fullImage.value ? '0' : 'auto');

const isActive = computed<boolean>(() => props.activeId === props.pageId);

const pageClasses = computed<string[]>(() => {
	const classes: string[] = [];
	if (isActive.value) classes.push('active');
	if (props.vertical) classes.push('vertical');
	if (props.img && !imageError.value) classes.push('has-image');
	if (fullImage.value) classes.push('full-image');
	if (props.imgObjectFit === ObjectFitValues.Cover) classes.push('image-cover');
	if (props.imgObjectFit === ObjectFitValues.Contain) classes.push('image-contain');
	if (props.imgObjectFit === ObjectFitValues.Fill) classes.push('image-fill');
	return classes;
});

function calcPageHeight(minHeight = 0): number {
	if (!dummy.value) return minHeight;

	const pageWidth = dummy.value.getBoundingClientRect().width;
	const pageContentInner = dummy.value.querySelector('.page-content-inner');
	let height = pageContentInner?.getBoundingClientRect().height || 0;

	if (!props.vertical && props.img && !imageError.value && fullImage.value) {
		const imgContainer = dummy.value.querySelector('.image-container') as HTMLDivElement;
		const img = dummy.value.querySelector('img.image') as HTMLImageElement;
		if (!imgContainer || !img || !img.complete) return minHeight;

		imgContainer.style.width = '0';
		for (let i = 0; i <= pageWidth; i++) {
			imgContainer.style.width = `${i}px`;
			const contentInnerHeight = pageContentInner?.getBoundingClientRect().height || 0;
			const containerHeight = imgContainer.getBoundingClientRect().height;
			const imgHeight = img.getBoundingClientRect().height;

			if (imgHeight >= containerHeight && contentInnerHeight <= containerHeight && containerHeight >= minHeight) {
				imageWidth.value = `${i}px`;
				height = containerHeight;
				imgContainer.style.width = '';

				return height;
			}
		}

		// fallback, if full image does not fit next to content
		imgContainer.style.width = '30%';
		const contentInnerHeight = pageContentInner?.getBoundingClientRect().height || 0;
		const containerHeight = imgContainer.getBoundingClientRect().height;
		const imgHeight = img.getBoundingClientRect().height;
		imageWidth.value = `${imgContainer.getBoundingClientRect().width}px`;
		imgContainer.style.width = '';
		height = Math.max(minHeight, contentInnerHeight, containerHeight, imgHeight);

		return height;
	}

	if (props.vertical) {
		const image = dummy.value.querySelector('.image-container');
		height += image?.getBoundingClientRect().height || 0;
	}

	return height;
}

function onImageLoad() {
	emit('imageLoaded', props.img!);
}

function onImageError() {
	imageError.value = true;
	emit('imageLoaded', props.img!);
}
</script>
<style scoped>
.page {
	@apply absolute left-0 bottom-0 size-full pr-8 grid grid-cols-1 gap-8 opacity-0 pointer-events-none min-h-full;

	.page-content {
		@apply flex flex-col justify-end flex-1;

		.page-content-inner {
			@apply py-8 w-full;
		}
	}

	.image-container {
		@apply flex items-center overflow-hidden;

		img.image {
			@apply w-full object-contain;
		}
	}

	&.full-image {
		@apply flex items-stretch;
	}

	&.vertical {
		@apply pr-0 gap-0 flex flex-col;

		.page-content {
			@apply flex-auto;

			.page-content-inner {
				@apply px-8 pt-0 pb-8;
			}
		}

		.image-container {
			@apply !w-full max-h-[50vh] flex justify-center;
		}
	}

	&.has-image {
		@apply grid-cols-2 pr-0;
	}

	&.image-cover {
		.image-container {
			img.image {
				@apply size-full object-cover;
			}
		}
	}

	&.image-contain {
		.image-container {
			img.image {
				@apply size-full object-contain;
			}
		}
	}

	&.image-fill {
		.image-container {
			img.image {
				@apply size-full object-fill;
			}
		}
	}

	&.active {
		@apply opacity-100 pointer-events-auto;
	}

	&.dummy {
		@apply min-h-0 h-auto pointer-events-none opacity-0;
	}
}
</style>
