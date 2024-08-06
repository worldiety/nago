import type {Ptr} from '@/shared/protocol/ora/ptr';
import type {ScopeID} from '@/shared/protocol/ora/scopeID';
import {inject} from 'vue';
import {uploadRepositoryKey} from '@/shared/injectionKeys';

export class UploadRepository {

	private readonly uploads = new Map<string, XMLHttpRequest>;

	fetchUpload(
		file: File,
		uploadId: string,
		receiverPtr: Ptr,
		scope: ScopeID,
		uploadProgressCallback: UploadProgressCallback,
		uploadFinishedCallback: UploadFinishedCallback,
		uploadAbortedCallback: UploadAbortedCallback,
		uploadFailedCallback: UploadFailedCallback,
	): Promise<void> {
		const formData = new FormData();
		formData.append(file.name, file, file.name);

		return new Promise<void>((resolve) => {
			const request = new XMLHttpRequest();
			request.upload.addEventListener('progress', (event: ProgressEvent) => {
				uploadProgressCallback(uploadId, event.loaded, event.total);
			});
			request.addEventListener('error', () => {
				uploadFailedCallback(uploadId, request.status);
				resolve();
			});
			request.addEventListener('load', () => {
				if (request.status.toString(10).startsWith('2')) {
					uploadFinishedCallback(uploadId);
					resolve();
					return;
				}
				uploadFailedCallback(uploadId, request.status);
				resolve();
			});
			request.addEventListener('abort', () => {
				uploadAbortedCallback(uploadId);
				resolve();
			})

			request.open('POST', '/api/ora/v1/upload');
			request.setRequestHeader("x-scope", scope)
			request.setRequestHeader("x-receiver", uploadId)
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
