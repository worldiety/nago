import { useHttp } from "@/shared/http";
import { PageConfiguration, UiDescription, UiEvent } from "@/shared/model";
import { useRoute } from "vue-router";
import { Ref } from "vue";

/**
 * A set of methods returned by {@link useUiEvents}.
 */
export interface UiEvents {
    /**
     * Send the event and update the current page.
     * @param event The event to send.
     */
    send: (event: UiEvent) => Promise<void>;
}

interface EventRequest {
    eventType: string,
    data: any,
    formData: Record<string, string>,
    model: any,
}

/**
 * Hook for sending events with all data of the current page.
 */
export function useUiEvents(ui: Ref<UiDescription>): UiEvents {
    const http = useHttp();
    const route = useRoute();
    const routeMeta = route.meta.page as PageConfiguration | undefined;
    if (!routeMeta) {
        throw new Error("useUiEvents not available on this page, as it not dynamically built from a PageConfiguration.");
    }

    const url = `http://localhost:3000${routeMeta.endpoint}`;

    async function send(event: UiEvent) {
        // TODO Collect all inputs from current page, by scraping the DOM.
        // const … = document.getElementsByName("input");
        const formData = {};

        let request: EventRequest = {
            data: event.data,
            eventType: event.eventType,
            formData,
            model: ui.value.viewModel,
        };
        
        const response = await http.request(url, "POST", request);
        ui.value = await response.json();
    }

    return { send };
}
