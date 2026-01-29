/**
 * Environment Configuration
 * Centralized access to environment variables with type safety
 */

export const env = {
  graphqlUrl: import.meta.env.VITE_GRAPHQL_URL || 'http://localhost:8080/graphql',
  wsUrl: import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws',
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  isDevelopment: import.meta.env.DEV,
  isProduction: import.meta.env.PROD,
} as const;
