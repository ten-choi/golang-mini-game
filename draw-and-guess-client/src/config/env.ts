/**
 * Environment Configuration
 * Centralized access to environment variables with type safety
 */

export const env = {
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  wsUrl: import.meta.env.VITE_WS_URL || 'ws://localhost:8080/app/ws',
  isDevelopment: import.meta.env.DEV,
  isProduction: import.meta.env.PROD,
} as const;

// Type for environment variables
declare global {
  interface ImportMetaEnv {
    VITE_API_BASE_URL?: string;
    VITE_WS_URL?: string;
  }
}
