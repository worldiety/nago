import type { CallServerFunc } from '@/shared/model/callServerFunc';
import type { SetServerProperty } from '@/shared/model/setServerProperty';
import type { UpdateJWT } from '@/shared/model/updateJWT';
import type { ClientHello } from '@/shared/model/clientHello';

export interface CallBatch {
	tx: (CallServerFunc | SetServerProperty | UpdateJWT | ClientHello) []
}
