# Query Keys Directory

This directory contains all the query keys used throughout the application for React Query. The query keys are organized by feature for better code splitting and maintainability.

## Structure

- `audiobooks.ts` - Query keys for audiobook-related operations
- `auth.ts` - Query keys for authentication-related operations
- `uploads.ts` - Query keys for file upload operations
- `index.ts` - Central export file for all query keys

## Usage

Import query keys from the central index file:

```typescript
import { audiobookKeys, authKeys, uploadKeys } from "@/queryKeys";
```

Or import specific query keys directly:

```typescript
import { audiobookKeys } from "@/queryKeys/audiobooks";
import { authKeys } from "@/queryKeys/auth";
import { uploadKeys } from "@/queryKeys/uploads";
```

## Benefits of Code Splitting

1. **Better Tree Shaking**: Only import the query keys you need
2. **Improved Bundle Size**: Reduces the overall bundle size by splitting code
3. **Better Maintainability**: Each feature has its own query key file
4. **Type Safety**: TypeScript ensures correct usage of query keys
5. **Centralized Management**: All query keys are in one place

## Query Key Patterns

Each query key follows a consistent pattern:

```typescript
export const featureKeys = {
  all: ["feature"] as const,
  lists: () => [...featureKeys.all, "list"] as const,
  list: (filters: string) => [...featureKeys.lists(), { filters }] as const,
  details: () => [...featureKeys.all, "detail"] as const,
  detail: (id: string) => [...featureKeys.details(), id] as const,
};
```

This pattern allows for:

- Invalidating all feature-related queries
- Invalidating specific list queries
- Invalidating specific detail queries
- Adding additional context to queries (like filters)
