import {HttpRequest} from "@/shared/http";

//TODO: Klasse anlegen?

export async function fetchUpload(files: Blob[], pageToken: string, uploadToken: string): Promise<void> {
	if (files.length === 0){
		console.log('Provided file array is empty, skipped upload!')
		return
	}
	const request = HttpRequest.post('/api/v1/upload')
	for (let i = files.length - 1; i >= 0; i--) {
		request.body(files[i])
	}
		request.header('x-page-token', pageToken)
		request.header('x-upload-token', uploadToken)
		await request.fetch()
}
