import type { InjectionKey } from 'vue';
import type EventBus from '@/shared/eventbus/eventBus';
import type ServiceAdapter from '@/shared/network/serviceAdapter';

export const eventBusKey = Symbol() as InjectionKey<EventBus>;
export const serviceAdapterKey = Symbol() as InjectionKey<ServiceAdapter>;
