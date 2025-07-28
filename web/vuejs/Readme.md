## VueJS Frontend Application

This folder contains a generic vuejs frontend application which is build and embedded into the nago framework.
It has the following core features:

* Tailwind-based ora components
* The api endpoint is the /api/... from the delivering http server
* All routes are requested once from GET /api/v1/nago/routes. It follows the semantic rules from https://github.com/golang/go/issues/61410 and not those of any javascript routing library.
```json
{
    "routes": [
      {
        "name": "Home",
        "path": "/"
      },
      {
        "name": "Product Details",
        "path": "/product/{productID}"
      }
    ]
  }
  ```

* Each call to a route requests its renderable component tree from /api/v1/nago/route/render/...
  * GET just renders the default registered model
  * POST triggers a re-rendering by submitting the last returned model and an optional ui.Event, which may also include form fields
  * The result is always a 200 JSON response, but of the following different types:
    * a render tree for the route
    * a redirect, either with forward or backward semantics. A forward pushes to the navigation stack and a backward removes all steps from history until the (old) route has been found and removed.
    * all values of named input types will be written into the event.

### Local development and build

This project uses **two separate `index.html` files**. `index.html` enables hot reloading for local development, `index-build.html` contains a script that determines the browser version via `bowser`
and dynamically loads one of **two Vite build targets** to support both **modern** and **legacy** browsers.

We generate **two builds**:

1. modern
	 * Output-Directory: `dist/modern/`
   * targets modern browsers with ECMAScript `esnext`
2. legacy
	 * Output-Directory: `dist/legacy/`
	 * targets legacy browsers with `Proxy` support
   * the generated build output by vite gets transpiled via babel (see `babel.config.json`) and loads three polyfill-scripts to provide functionalities for a wide range of browsers
     * https://cdn.jsdelivr.net/npm/es5-polyfill@0.0.6/dist/polyfill.min.js
     * https://cdn.jsdelivr.net/npm/core-js-bundle@3/minified.js
     * https://cdn.jsdelivr.net/npm/regenerator-runtime/runtime.js

After the browser is identified by `bowser`, we use the thresholds defined in `index-build.html` to select either the legacy or modern bundle.
