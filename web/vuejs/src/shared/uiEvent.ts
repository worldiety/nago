import router from '@/router';
import { useHttp } from '@/shared/http';
import type { PageConfiguration, UiDescription, UiEvent } from '@/shared/model';
import type { Ref } from 'vue';
import { useRoute } from 'vue-router';

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
    // EventType is the absolute qualified type name as it was defined within the render tree.
    eventType: string;
    // EventData is exactly the serialized payload of the Event which has been defined within the render tree.
    eventData: any;
    // FormData is whatever the client wants to send, e.g. input text data, options or even file uploads.
    formData: any;
    // Model is whatever the server has used to build the render tree. This allows keeping the server stateless so far.
    model: any;
}

interface FileData {
    name: string;
    lastModified: number;
    size: number;
    type: string;
    data: string; // base64
}

/**
 * Hook for sending events with all data of the current page.
 */
export function useUiEvents(ui: Ref<UiDescription>): UiEvents {
    const http = useHttp();
    const route = useRoute();
    const routeMeta = route.meta.page as PageConfiguration | undefined;
    if (!routeMeta) {
        throw new Error(
            'useUiEvents not available on this page, as it not dynamically built from a PageConfiguration.'
        );
    }

    const url = `http://localhost:3000${routeMeta.endpoint}`;

    async function send(event: UiEvent) {
        if (event == null) {
            console.log('send event ignored: null');
            return new Promise<void>((resolve, reject) => resolve());
        }

        const formData = {};
        const inputElems = document.getElementsByTagName('input');
        for (let i = 0; i < inputElems.length; i++) {
            const item = inputElems.item(i);
            if (item == null) {
                return;
            }

            const name = item.getAttribute('name');
            if (name == null || name == '') {
                continue;
            }

            if (item.getAttribute('type') === 'file') {
                if (item.files == null) {
                    continue;
                }

                if (item.multiple) {
                    const files = item.files;
                    const tmp = [];
                    for (let i = 0; i < files.length; i++) {
                        const file = files.item(i);
                        if (file == null) {
                            throw new Error('cannot happen!?');
                        }

                        const b64 = await readFileAsDataURL(file);
                        const fd: FileData = {
                            data: b64,
                            lastModified: file.lastModified,
                            name: file.name,
                            size: file.size,
                            type: file.type,
                        };
                        tmp.push(fd);
                    }
                    formData[name] = tmp;
                } else {
                    const files = item.files;
                    if (files.length > 0) {
                        const file = files[0];
                        const b64 = await readFileAsDataURL(file);
                        const fd: FileData = {
                            data: b64,
                            lastModified: file.lastModified,
                            name: file.name,
                            size: file.size,
                            type: file.type,
                        };

                        formData[name] = fd;
                    }
                }
            } else {
                // todo don't know how to let js and ts world be fine together here
                formData[name] = item.value;
            }
        }

        // console.log(formData) // file upload will bring the browser in debug print down
        console.log('form data collected and POST');

        const request: EventRequest = {
            eventData: event.data,
            eventType: event.eventType,
            formData,
            model: ui.value.viewModel,
        };

        const response = await http.request(url, 'POST', request);
        ui.value = await response.json();

        if (ui.value?.redirect?.redirect) {
            //console.log("shall redirect")
            //console.log(ui.value?.redirect)
            await router.push({ path: ui.value.redirect.url }).catch((e) => console.log(e));
            // TODO @Lars ich bekomme die Navigation und die vor-zurück Historie nicht sauber hin
            //window.location.reload()
            router.go(0);
            return new Promise<void>((ok, no) => ok());
        }
    }

    return { send };
}

async function readFileAsDataURL(file: File) {
    let resultBase64 = await new Promise<string>((resolve) => {
        const fileReader = new FileReader();
        // TODO i don't understand why the linter is unhappy here, the doc says its always a string?
        fileReader.onload = (e) => resolve(fileReader.result);
        fileReader.readAsDataURL(file);
    });

    resultBase64 = resultBase64.split(',')[1];
    return resultBase64;
}
