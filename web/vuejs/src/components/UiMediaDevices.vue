<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import type { MediaDevices } from '@/shared/proto/nprotoc_gen';
import { MediaDevice, MediaDeviceKindValues } from '@/shared/proto/nprotoc_gen';
import { UpdateStateValueRequested } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: MediaDevices;
}>();

const serviceAdapter = useServiceAdapter();
const currentMediaDevices = ref<MediaDevice[]>([]);
const hasCameraPermission = ref(false);

async function checkCameraPermission() {
	try {
		await navigator.mediaDevices.getUserMedia({
			audio: props.ui.withAudio ?? false,
			video: props.ui.withVideo ?? false,
		});
		hasCameraPermission.value = true;
	} catch (e) {
		hasCameraPermission.value = false;
		console.warn("Couldn't get requested permissions", e);
	}
	serviceAdapter.sendEvent(
		new UpdateStateValueRequested(props.ui.hasGrantedPermissions, 0, nextRID(), String(hasCameraPermission.value))
	);
}
async function getMediaDevices() {
	const devices = await navigator.mediaDevices.enumerateDevices();
	currentMediaDevices.value = devices.filter(isRequested).map(mapMediaDeviceInfoToMediaDevice);
	const payload = JSON.stringify(currentMediaDevices.value);
	serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), payload));
}

function getMediaDeviceKindFromMediaDeviceInfo(device: MediaDeviceInfo): MediaDeviceKindValues {
	switch (device.kind) {
		case 'audioinput':
			return MediaDeviceKindValues.AudioInput;
		case 'audiooutput':
			return MediaDeviceKindValues.AudioOutput;
		case 'videoinput':
			return MediaDeviceKindValues.VideoInput;
	}
}

function isRequested(device: MediaDeviceInfo): boolean {
	if (!device.label || !device.deviceId || !device.groupId) return false;

	switch (device.kind) {
		case 'audioinput':
		case 'audiooutput':
			return props.ui.withAudio ?? false;
		case 'videoinput':
			return props.ui.withVideo ?? false;
	}
}

function mapMediaDeviceInfoToMediaDevice(device: MediaDeviceInfo): MediaDevice {
	return new MediaDevice(
		device.deviceId,
		device.groupId,
		device.label,
		getMediaDeviceKindFromMediaDeviceInfo(device)
	);
}

onMounted(async () => {
	await checkCameraPermission();
	if (hasCameraPermission.value) {
		await getMediaDevices();
	}
});
</script>

<template>
	<div></div>
</template>
