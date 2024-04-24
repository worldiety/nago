import type { InjectionKey } from 'vue';
import type EventBus from '@/shared/eventBus';

export const eventBusKey = Symbol() as InjectionKey<EventBus>;
