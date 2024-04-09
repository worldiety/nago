import type {Pointer} from "@/shared/protocol/pointer";

/**
 * A Property represents a unique field or attribute within a backends scope.
 * A component usually consists of at least a meaningful property (e.g. a text content).
 * The generic T of a property can be a primitive but also something complex like another
 * Component.
 */
export interface Property<T> {
	/**
	 * p is short for "Pointer" and references a property instance within the backend.
	 *
	 */
	p: Pointer;
	/**
	 * v contains the actual value specified by the generic type parameter.
	 */
	v: T;
}
