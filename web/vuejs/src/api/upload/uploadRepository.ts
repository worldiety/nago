//TODO: Klasse anlegen?

export type UploadProgressCallback = (progress: number, total: number) => void;

export async function fetchUpload(files: File[], pageToken: string, uploadToken: string, uploadProgressCallback: UploadProgressCallback): Promise<void> {
	if (files.length === 0) {
		return;
	}
	const formData = new FormData();
	files.forEach((file: File) => {
		formData.append(file.name, file, file.name);
	});

	return new Promise<void>((resolve, reject) => {
		const request = new XMLHttpRequest();
		request.upload.addEventListener('progress', (event: ProgressEvent) => {
			uploadProgressCallback(event.loaded, event.total);
		});
		request.addEventListener('error', (e) => {
			console.log('ERR', e);
			reject('Error');
		});
		request.addEventListener('load', () => {
			request.status.toString(10).startsWith('2') ? resolve() : reject(request.status);
		});
		request.addEventListener('abort', () => {
			console.log('ABORTED');
			reject('Aborted');
		})
		request.open('POST', '/api/v1/upload');
		request.send(formData);
	});
}
