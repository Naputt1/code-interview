import { defineConfig } from 'orval'

/**
 * Generates the typed TanStack Query hooks for the whole app from the backend's
 * OpenAPI spec (written by `make swagger` from the Go handler annotations).
 * Every request goes through `customInstance` so error shaping lives in one
 * place.
 */
export default defineConfig({
  api: {
    input: '../backend/api/swagger/openapi3.yaml',
    output: {
      client: 'react-query',
      target: './src/api/gen/index.ts',
      mode: 'single',
      mock: false,
      clean: true,
      override: {
        query: {
          version: 5,
        },
        fetch: {
          includeHttpResponseReturnType: false,
        },
        mutator: {
          path: './src/api/mutator.ts',
          name: 'customInstance',
        },
      },
    },
  },
})
