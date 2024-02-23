import axios from 'axios';

export async function fetchBackendData(): Promise<string[]> {
    try {
        const response = await axios.get('http://localhost:8090/api/v1/ui/application');
        const data = response.data;
        const anchors = data.livePages.map((page: { anchor: string }) => page.anchor);
        return anchors;
    } catch (error) {
        console.error('Error fetching backend data:', error);
        return [];
    }
}
