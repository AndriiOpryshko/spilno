import { defineConfig } from 'orval';

// The API contract lives in the core service; this regenerates src/api/gen from it.
export default defineConfig({
  core: {
    input: { target: '../../services/core/api/openapi.yaml' },
    output: {
      mode: 'tags-split',
      target: 'src/api/gen/endpoints',
      schemas: 'src/api/gen/model',
      client: 'react-query',
      httpClient: 'fetch',
      // The dev server and preview proxy /api to the core service.
      baseUrl: '/api',
      clean: true,
      override: {
        // Throw non-2xx responses so TanStack Query sees them as errors typed with Problem.
        fetch: { forceSuccessResponse: true },
      },
    },
  },
});
