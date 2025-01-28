FileImportRequested asks the frontend to let the user pick some files.
Depending on the actual backend configuration, this may cause
a regular http multipart upload or some FFI calls providing data streams
or accessor URIs.