/**
 * Code generated by github.com/worldiety/macro. DO NOT EDIT.
 */


import type { EventType } from '@/shared/protocol/ora/eventType';

/**
 * FileImportRequested asks the frontend to let the user pick some files.
 * Depending on the actual backend configuration, this may cause
 * a regular http multipart upload or some FFI calls providing data streams
 * or accessor URIs.
 */
export interface FileImportRequested {
    // Type
    type: 'FileImportRequested'/*EventType*/;
    // ID
    id /*ID*/: string;
    // ScopeID
    scopeID: string;
    // Multiple
    multiple: boolean;
    // MaxBytes
    maxBytes: number /*int64*/;
    // AllowedMimeTypes
    allowedMimeTypes: string[];
}

