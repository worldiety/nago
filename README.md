# nago

## Authentication Demo

To run the authentication demo you need to have docker-compose, npm and go installed.

We will be using Keycloak as an IDP, although any OIDC compatible IDP would work. The Keycloak service is defined in `compose.yaml` and can be launched in the background with

```shell
docker-compose up --detach
```

Some settings need to be configured in Keycloak:
* Open <http://localhost:8080> and sign in as admin/admin (as defined in `compose.yaml`)
* Create a new realm "nago" using the dropdown in the top-left corner
* You might want to allow users to register accounts by themselves under Realm settings &rArr; Login &rArr; User registration. You can enable "Email as username" here to stop worrying about usernames.
* Create a new client
  * Set the client ID to "nago"
  * Make sure the client type is OpenID Connect
  * Make sure the "Standard Flow" is enabled.
  * Set Root URL to "http://localhost:5173"
  * Set Valid redirect URIs to "/oauth"
  * Set Valid post logout redirect URIs to "http://localhost:5173"
  * Set Web origins to "http://localhost:5173"

With the Keycloak configuration complete, we can now launch the frontend and backend.

Install npm dependencies and launch the frontend with:

```shell
cd web/vuejs
npm install
npm run dev
```

Then launch the backend:

```shell
go run ./example/cmd/auth-demo
```

You can now use the frontend at <http://localhost:5173>. Make sure to have a look at `web/vuejs/src/stores/authStore.ts` to see how the `oidc-client-ts` library is configured. Check the files in `web/vuejs/src/views` to see how authentication is used in the application. To have authentication checks on a per-route basis, take a look at `web/vuejs/src/router/index.ts`.
