/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { inject } from 'vue';
import { uploadProgressManager } from '@/api/upload/uploadProgressManager';
import { uploadRepositoryKey } from '@/shared/injectionKeys';
import { Ptr, ScopeID } from '@/shared/proto/nprotoc_gen';

export class UploadRepository {
	private readonly uploads = new Map<string, XMLHttpRequest>();

	fetchUpload(
		file: File,
		uploadId: string,
		receiverPtr: Ptr,
		scope: ScopeID,
		uploadProgressCallback: UploadProgressCallback,
		uploadFinishedCallback: UploadFinishedCallback,
		uploadAbortedCallback: UploadAbortedCallback,
		uploadFailedCallback: UploadFailedCallback
	): Promise<void> {
		const formData = new FormData();
		formData.append(file.name, file, file.name);

		uploadProgressManager.addUpload(uploadId, file.name, file.size);

		return new Promise<void>((resolve) => {
			const request = new XMLHttpRequest();
			request.upload.addEventListener('progress', (event: ProgressEvent) => {
				uploadProgressCallback(uploadId, event.loaded, event.total);

				const percent = Math.round((event.loaded / event.total) * 100);
				uploadProgressManager.updateProgress(uploadId, percent);
			});
			request.addEventListener('error', () => {
				uploadProgressManager.removeUpload(uploadId);
				uploadFailedCallback(uploadId, request.status);
				resolve();
			});
			request.addEventListener('load', () => {
				uploadProgressManager.removeUpload(uploadId);
				if (request.status.toString(10).startsWith('2')) {
					uploadFinishedCallback(uploadId);
					resolve();
					return;
				}
				uploadFailedCallback(uploadId, request.status);
				resolve();
			});
			request.addEventListener('abort', () => {
				uploadProgressManager.removeUpload(uploadId);
				uploadAbortedCallback(uploadId);
				resolve();
			});

			request.open('POST', '/api/ora/v1/upload');
			request.setRequestHeader('x-scope', scope);
			request.setRequestHeader('x-receiver', uploadId);
			request.send(formData);
			this.uploads.set(uploadId, request);
		});
	}

	abortUpload(uploadId: string): void {
		this.uploads.get(uploadId)?.abort();
		this.uploads.delete(uploadId);
	}
}

export type UploadProgressCallback = (uploadId: string, progress: number, total: number) => void;

export type UploadFinishedCallback = (uploadId: string) => void;

export type UploadAbortedCallback = (uploadId: string) => void;

export type UploadFailedCallback = (uploadId: string, statusCode: number) => void;

export function useUploadRepository(): UploadRepository {
	const uploadRepository = inject(uploadRepositoryKey);
	if (!uploadRepository) {
		throw new Error('Could not inject UploadRepository as it is undefined');
	}

	return uploadRepository;
}
