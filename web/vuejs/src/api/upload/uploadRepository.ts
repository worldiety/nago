
/*
import {useHttp} from "@/shared/http";


export function fileInputChange() {
	const item = e.target;
	const formData = new FormData();
	for (const file of item.files) {
		formData.append('files', file);
	}



	//TODO: auslagern
	// - Upload Repository anlegen, das diesen Endpunkt abfragt /api/v1/upload
	// - Überprüfen, ob ich den Fehler an dieser Stelle anzeigen kann, ansonsten in App.vue durchreichen
	// -
	// -
	fetch('/api/v1/upload', {
		method: 'POST',
		body: formData,
		/*
		headers: {
			'x-page-token': props.page.token,
			'x-upload-token': props.ui.uploadToken.value,
		},


	});
}



 */
