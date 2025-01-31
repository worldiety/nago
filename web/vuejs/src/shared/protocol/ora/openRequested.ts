/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */
import type { EventType } from '@/shared/protocol/ora/eventType';

/**
 * NavigationBackRequested steps back causing a likely destruction of the most top component.
 * The frontend may deproto.Ptre to ignore that, if the stack would be empty/undefined otherwise.
 */
export interface OpenRequested {
	// Type
	type: 'OpenRequested' /*EventType*/;
	resource: string;
	options: Record<string, string>;
}
