<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
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

const error = ref<string>();

const frameStyles = computed<string>(() => {
	const styles = frameCSS(props.ui.frame);

	return styles.join(';');
});
const constraints = computed<MediaTrackConstraints>(() => {
	if (!props.ui.mediaDevice) return { deviceId: 'disabled' };

	return { deviceId: props.ui.mediaDevice.deviceID, groupId: props.ui.mediaDevice.groupID };
});

function submitValue(newValue: string) {
	// Note, that the sendEvent may have a huge latency, causing ghost updates for the user input.
	// Thus, immediately increase the request id, so that everybody knows, that any older responses are outdated.
	nextRID();

	serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), newValue));
}
function onDetect(detectedCodes: DetectedBarcode[]) {
	const result = JSON.stringify(detectedCodes.map((code) => code.rawValue));

	submitValue(result);
}
function onCameraReady() {
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
</script>

<template>
	<div :style="frameStyles">
		<qrcode-stream
			:track="paintBoundingBox"
			:formats="['qr_code']"
			:constaints="constraints"
			:torch="props.ui.activatedTorch"
			@detect="onDetect"
			@error="onError"
			@camera-on="onCameraReady"
		></qrcode-stream>
	</div>
</template>
