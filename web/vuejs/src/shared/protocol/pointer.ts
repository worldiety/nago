/**
 * A Pointer represents a remote object within a Scope inside the backend.
 * Note, that you can guess other allocated properties, components or function pointers
 * within the same Scope (they are just incrementing over time).
 * However, that is not a special security problem, because these are only visual things
 * and it is just as secure as it would be, when navigating around by hand.
 * Secrets must never reside inside "hidden" properties.
 *
 * You cannot break into foreign scopes, you need to guess the Scope identifier to hijack it - just like
 * a conventional session id.
 */
export type Pointer = number
