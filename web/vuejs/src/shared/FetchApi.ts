const TIMEOUT_MILLISECONDS = 10000;

const get = async (path: string): Promise<unknown> => {
    const url = path;
    const response = await fetchWithTimeout(url, {
        method: 'GET',
    });

    if (response.ok) {
        return await response.json();
    } else {
        handleDefectiveResponse(response, url);
    }
};

const getWithoutResponse = async (path: string): Promise<void> => {
    const url = path;
    const response = await fetchWithTimeout(url, {
        method: 'GET',
    });

    if (!response.ok) {
        handleDefectiveResponse(response, url);
    }
};

const post = async (path: string, request: unknown): Promise<void> => {
    const url = path;
    const response = await fetchWithTimeout(url, {
        method: 'POST',
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    });

    if (!response.ok) {
        handleDefectiveResponse(response, url);
    }
};

const postWithResponse = async (path: string, request: unknown): Promise<unknown> => {
    const url = path;
    const response = await fetchWithTimeout(url, {
        method: 'POST',
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    });

    if (response.ok) {
        return await response.json();
    } else {
        handleDefectiveResponse(response, url);
    }
};

const put = async (path: string, request: unknown): Promise<void> => {
    const url = path;
    const response = await fetchWithTimeout(url, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    });

    if (!response.ok) {
        handleDefectiveResponse(response, url);
    }
};

async function fetchWithTimeout(url: string, options: RequestInit) {
    const controller = new AbortController();
    const id = setTimeout(() => controller.abort(), TIMEOUT_MILLISECONDS);

    const response = await fetch(url, {
        ...options,
        signal: controller.signal,
    });
    clearTimeout(id);

    return response;
}

function handleDefectiveResponse(response: Response, url: string) {
    const errorMessage = 'Error calling ' + url;
    console.error(errorMessage);
    throw new Error(errorMessage);
}

export { get, getWithoutResponse, post, postWithResponse, put };
