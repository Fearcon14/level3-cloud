# paas-ui

This template should help get you started developing with Vue 3 in Vite.

## Recommended IDE Setup

[VS Code](https://code.visualstudio.com/) + [Vue (Official)](https://marketplace.visualstudio.com/items?itemName=Vue.volar) (and disable Vetur).

## Recommended Browser Setup

- Chromium-based browsers (Chrome, Edge, Brave, etc.):
  - [Vue.js devtools](https://chromewebstore.google.com/detail/vuejs-devtools/nhdogjmejiglipccpnnnanhbledajbpd)
  - [Turn on Custom Object Formatter in Chrome DevTools](http://bit.ly/object-formatters)
- Firefox:
  - [Vue.js devtools](https://addons.mozilla.org/en-US/firefox/addon/vue-js-devtools/)
  - [Turn on Custom Object Formatter in Firefox DevTools](https://fxdx.dev/firefox-devtools-custom-object-formatters/)

## Customize configuration

See [Vite Configuration Reference](https://vite.dev/config/).

## Project Setup

```sh
npm install
```

### Compile and Hot-Reload for Development

```sh
npm run dev
```

### Compile and Minify for Production

```sh
npm run build
```

### Run Unit Tests with [Vitest](https://vitest.dev/)

```sh
npm run test:unit
```

### Run End-to-End Tests with [Playwright](https://playwright.dev)

```sh
# Install browsers for the first run
npx playwright install

# When testing on CI, must build the project first
npm run build

# Runs the end-to-end tests
npm run test:e2e
# Runs the tests only on Chromium
npm run test:e2e -- --project=chromium
# Runs the tests of a specific file
npm run test:e2e -- tests/example.spec.ts
# Runs the tests in debug mode
npm run test:e2e -- --debug
```

### Lint with [ESLint](https://eslint.org/)

```sh
npm run lint
```

## Deployment (SKE / Argo CD)

The UI is containerized and deployed via Argo CD. The same ingress host serves the UI at `/` and the API at `/api`, so the frontend uses relative URLs and needs no API base URL in production.

### Build and push image to STACKIT Container Registry

```sh
# From Week5_Frontend/paas-ui
docker build --platform linux/amd64 -t registry.onstackit.cloud/kevin-sinn/paas-ui:latest .
docker push registry.onstackit.cloud/kevin-sinn/paas-ui:latest
```

Ensure you are logged in to the registry (e.g. `docker login registry.onstackit.cloud`) and that the `stackit-registry` image pull secret exists in the cluster (same as for the API).

### Local development and API backend

For local dev, API requests are proxied by Vite. By default the proxy target is `http://localhost:8080`. To point at your cluster’s API (e.g. LoadBalancer IP), set:

```sh
export VITE_API_PROXY_TARGET=http://<YOUR_LB_IP>
npm run dev
```

See `.env.development.example` for an example.
