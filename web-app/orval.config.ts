import { defineConfig } from 'orval'

export default defineConfig({
  auth: {
    input: {
      target: './openapi/auth.yaml',
    },
    output: {
      target: './src/shared/api/generated/auth.ts',
      client: 'fetch',
      baseUrl: '/api/v1',
      override: {
        mutator: {
          path: './src/shared/api/instance.ts',
          name: 'customFetch',
        },
      },
    },
  },
  users: {
    input: {
      target: './openapi/users.yaml',
    },
    output: {
      target: './src/shared/api/generated/users.ts',
      client: 'fetch',
      baseUrl: '/api/v1',
      override: {
        mutator: {
          path: './src/shared/api/instance.ts',
          name: 'customFetch',
        },
      },
    },
  },
  tasks: {
    input: {
      target: './openapi/tasks.yaml',
    },
    output: {
      target: './src/shared/api/generated/tasks.ts',
      client: 'fetch',
      baseUrl: '/api/v1',
      override: {
        mutator: {
          path: './src/shared/api/instance.ts',
          name: 'customFetch',
        },
      },
    },
  },
})
