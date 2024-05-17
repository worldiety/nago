//TODO: Klasse anlegen?

import type { Ptr } from '@/shared/protocol/ora/ptr';
import type { ScopeID } from '@/shared/protocol/ora/scopeID';

export type UploadProgressCallback = (uploadId: string, progress: number, total: number) => void;

export type UploadFinishedCallback = (uploadId: string) => void;

export async function fetchUpload(
		file: File,
		uploadId: string,
		receiverPtr: Ptr,
		scope: ScopeID,
		uploadProgressCallback: UploadProgressCallback,
		uploadFinishedCallback: UploadFinishedCallback,
	): Promise<void> {
	const formData = new FormData();
	formData.append(file.name, file, file.name);

	return new Promise<void>((resolve, reject) => {
		const request = new XMLHttpRequest();
		request.upload.addEventListener('progress', (event: ProgressEvent) => {
			uploadProgressCallback(uploadId, event.loaded, event.total);
		});
		request.addEventListener('error', (e) => {
			console.log('ERR', e);
			reject('Error');
		});
		request.addEventListener('load', () => {
			if (request.status.toString(10).startsWith('2')) {
				uploadFinishedCallback(uploadId);
				resolve();
				return;
			}
			reject(request.status);
		});
		request.addEventListener('abort', () => {
			console.log('ABORTED');
			reject('Aborted');
		})

		request.open('POST', '/api/ora/v1/upload');
		request.setRequestHeader("x-scope", scope)
		request.setRequestHeader("x-receiver", String(receiverPtr))
		request.send(formData);
	});
}
