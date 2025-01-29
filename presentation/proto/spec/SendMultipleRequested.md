SendMultipleRequested is an event for the frontend from the backend
to send the according resources into the system environment.
A Webbrowser may issue a regular download. A backend should not issue multiple downloads at once but instead
pack multiple files into a zip file because the browser support for something like a multipart download
is just broken today. An Android App may trigger the according Intent and opens a picker
to select the receiving app.