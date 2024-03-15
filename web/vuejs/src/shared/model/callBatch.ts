export interface CallBatch {
	tx: (CallServerFunc | SetServerProperty | UpdateJWT | ClientHello) []
}
