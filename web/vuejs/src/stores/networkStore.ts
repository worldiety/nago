import { defineStore } from 'pinia';
import WebSocketAdapter from '@/shared/webSocketAdapter';
import type { LiveMessage } from '@/shared/model/liveMessage';
import type { PropertyFunc } from '@/shared/model/propertyFunc';
import type { Property } from '@/shared/model/property';

interface NetworkStoreState {
  webSocketAdapter: WebSocketAdapter;
}

export const useNetworkStore = defineStore('networkStore', {
  state: (): NetworkStoreState => ({
    webSocketAdapter: new WebSocketAdapter(),
  }),
  actions: {
    initializeWebSocket(): void {
      this.webSocketAdapter.initializeWebSocket();
    },
    setWebSocketReceiveCallback(callback: (message: LiveMessage) => void): void {
      this.webSocketAdapter.setWebSocketReceiveCallback(callback);
    },
    setWebSocketErrorCallback(callback: () => void): void {
      this.webSocketAdapter.setWebSocketErrorCallback(callback);
    },
    closeWebSocket(): void {
      this.webSocketAdapter.closeWebSocket();
    },
    invokeFunc(action: PropertyFunc): void {
      this.webSocketAdapter.invokeFunc(action);
    },
    invokeSetProp(property: Property): void {
      this.webSocketAdapter.invokeSetProp(property);
    },
    invokeFuncAndSetProp(property: Property, action: PropertyFunc): void {
      this.webSocketAdapter.invokeFuncAndSetProp(property, action);
    },
  },
});
