<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import LoadingAnimation from '@/components/shared/LoadingAnimation.vue';
import { colorValue } from '@/components/shared/colors';
import { frameCSS } from '@/components/shared/frame';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import type { DetectedBarcode } from 'vue-qrcode-reader';
import { QrcodeStream } from 'vue-qrcode-reader';
import { FunctionCallRequested, QrCodeReader } from '@/shared/proto/nprotoc_gen';
import { UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';
import { colorToHexValue } from '@/shared/tailwindTranslator';

const props = defineProps<{
	ui: QrCodeReader;
}>();

const serviceAdapter = useServiceAdapter();

const isCameraReady = ref<boolean>(false);
const error = ref<string>();

const frameStyles = computed<string>(() => {
	const styles = frameCSS(props.ui.frame);

	return styles.join(';');
});

const constraints = computed<MediaTrackConstraints>(() => {
	return { deviceId: props.ui.mediaDevice?.deviceID, groupId: props.ui.mediaDevice?.groupID };
});

function onDetect(detectedCodes: DetectedBarcode[]) {
	const payload = JSON.stringify(detectedCodes.map((code) => code.rawValue));

	// Note, that the sendEvent may have a huge latency, causing ghost updates for the user input.
	// Thus, immediately increase the request id, so that everybody knows, that any older responses are outdated.
	nextRID();

	serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), payload));
}
function onCameraReady() {
	isCameraReady.value = true;
	serviceAdapter.sendEvent(new FunctionCallRequested(props.ui.onCameraReady, nextRID()));
}
function paintBoundingBox(detectedCodes: DetectedBarcode[], ctx: CanvasRenderingContext2D) {
	for (const detectedCode of detectedCodes) {
		const {
			boundingBox: { x, y, width, height },
		} = detectedCode;

		ctx.lineWidth = props.ui.trackerLineWidth ?? 2;
		const trackerColor = colorValue(props.ui.trackerColor) || colorValue('M0');
		ctx.strokeStyle = colorToHexValue(trackerColor);
		ctx.strokeRect(x, y, width, height);
	}
}
function onError(err: Error) {
	error.value = `[${err.name}]: `;

	if (err.name === 'NotAllowedError') {
		error.value += 'you need to grant camera access permission';
	} else if (err.name === 'NotFoundError') {
		error.value += 'no camera on this device';
	} else if (err.name === 'NotSupportedError') {
		error.value += 'secure context required (HTTPS, localhost)';
	} else if (err.name === 'NotReadableError') {
		error.value += 'is the camera already in use?';
	} else if (err.name === 'OverconstrainedError') {
		error.value += 'installed cameras are not suitable';
	} else if (err.name === 'StreamApiNotSupportedError') {
		error.value += 'Stream API is not supported in this browser';
	} else if (err.name === 'InsecureContextError') {
		error.value += 'Camera access is only permitted in secure context. Use HTTPS or localhost rather than HTTP.';
	} else {
		error.value += err.message;
	}

	console.error(error.value);
}
function onUpdatedConstraints(newConstraints?: MediaTrackConstraints, oldConstraints?: MediaTrackConstraints) {
	if (
		newConstraints &&
		oldConstraints &&
		newConstraints.deviceId === oldConstraints.deviceId &&
		newConstraints.groupId === oldConstraints.groupId
	)
		return;

	// if the constraints changed, we need to initialize a new camera, therefore enable the loading animation by setting isCameraReady to false
	isCameraReady.value = false;
}

watch(
	() => constraints.value,
	(newValue, oldValue) => onUpdatedConstraints(newValue, oldValue)
);
</script>

<template>
	<div class="relative" :style="frameStyles">
		<div v-if="props.ui.mediaDevice" class="relative h-full w-full">
			<qrcode-stream
				v-show="isCameraReady"
				:track="paintBoundingBox"
				:formats="['qr_code']"
				:constraints="constraints"
				:torch="props.ui.activatedTorch"
				@detect="onDetect"
				@error="onError"
				@camera-on="onCameraReady"
				@camera-off="isCameraReady = false"
			></qrcode-stream>
			<div v-if="!isCameraReady" class="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2">
				<LoadingAnimation class="h-5 w-5" />
			</div>
		</div>
		<div v-else class="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-full">
			<UiGeneric v-if="props.ui.noMediaDeviceContent" :ui="props.ui.noMediaDeviceContent" />
		</div>
	</div>
</template>
