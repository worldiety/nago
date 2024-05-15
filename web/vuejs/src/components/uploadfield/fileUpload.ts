export default interface FileUpload {
	uploadId: string;
	file: File;
	bytesUploaded: number|null;
	bytesTotal: number|null;
	finished: boolean;
}
