import type { InjectionKey } from 'vue';
import type EventBus from '@/shared/eventbus/eventBus';
import type ServiceAdapter from '@/shared/network/serviceAdapter';
import type { UploadRepository } from '@/api/upload/uploadRepository';
import type ThemeManager from '@/shared/themeManager';

export const eventBusKey = Symbol() as InjectionKey<EventBus>;
export const serviceAdapterKey = Symbol() as InjectionKey<ServiceAdapter>;
export const uploadRepositoryKey = Symbol() as InjectionKey<UploadRepository>;
export const themeManagerKey = Symbol() as InjectionKey<ThemeManager>;
