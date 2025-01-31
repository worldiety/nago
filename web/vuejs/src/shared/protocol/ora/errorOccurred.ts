/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */
import type { EventType } from '@/shared/protocol/ora/eventType';
import type { RequestId } from '@/shared/protocol/ora/requestId';

export interface ErrorOccurred {
	// Type
	type: 'ErrorOccurred' /*EventType*/;
	// RequestId
	r /*RequestId*/ : RequestId;
	// Message
	message: string;
}
