import type { CallServerFunc } from '@/shared/model/callServerFunc';
import type { ClientHello } from '@/shared/model/clientHello';
import type { SetServerProperty } from '@/shared/model/setServerProperty';
import type { UpdateJWT } from '@/shared/model/updateJWT';

/**
 * @deprecated use EventsAggregated
 */
export interface CallBatch {
	tx: (CallServerFunc | SetServerProperty | UpdateJWT | ClientHello)[];
}
