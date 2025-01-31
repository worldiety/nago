export interface FileUpload {
	uploadId: string;
	file: File;
	bytesUploaded: number | null;
	bytesTotal: number | null;
	status: FileUploadStatus;
	statusCode?: number;
}

export enum FileUploadStatus {
	PENDING,
	IN_PROGRESS,
	SUCCESS,
	ABORTED,
	ERROR,
}
