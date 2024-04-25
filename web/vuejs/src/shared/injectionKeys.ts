import type { InjectionKey } from 'vue';
import type EventBus from '@/shared/eventBus';
import ServiceAdapter from '@/shared/network/serviceAdapter';

export const eventBusKey = Symbol() as InjectionKey<EventBus>;
export const networkAdapterKey = Symbol() as InjectionKey<ServiceAdapter>;
