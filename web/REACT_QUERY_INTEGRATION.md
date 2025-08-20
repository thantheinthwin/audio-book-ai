# React Query Integration

This document describes the React Query (TanStack Query) integration in the Audio Book AI application.

## Overview

React Query has been integrated to provide:

- Efficient data fetching and caching
- Automatic background refetching
- Optimistic updates
- Error handling
- Loading states
- Request deduplication

## Setup

### Dependencies

```bash
npm install @tanstack/react-query @tanstack/react-query-devtools
```

### Configuration

The React Query client is configured in `lib/react-query.ts` with the following defaults:

- **Stale Time**: 5 minutes
- **GC Time**: 10 minutes
- **Retry Logic**: 3 attempts for server errors, no retry for 4xx errors
- **Refetch on Window Focus**: Disabled

### Provider Setup

The `QueryProvider` component wraps the entire application in `app/layout.tsx` and includes the React Query DevTools for development.

## Custom Hooks

### Authentication Hooks (`hooks/use-auth.ts`)

#### `useSession()`

Fetches the current user session from Supabase.

#### `useUser()`

Fetches the current user data from Supabase.

#### `useTestAuth()`

Mutation hook for testing authentication.

#### `useTestAdminAuth()`

Mutation hook for testing admin authentication.

#### `useSignOut()`

Mutation hook for signing out users.

### Audiobook Hooks (`hooks/use-audiobooks.ts`)

#### `useAudioBooks()`

Fetches all audiobooks from the API.

#### `useAudioBook(id: string)`

Fetches a specific audiobook by ID.

#### `useCreateAudioBook()`

Mutation hook for creating new audiobooks.

#### `useUpdateAudioBook()`

Mutation hook for updating existing audiobooks.

#### `useDeleteAudioBook()`

Mutation hook for deleting audiobooks.

#### `useAudioBookChapters(audiobookId: string)`

Fetches chapters for a specific audiobook.

### Upload Hooks (`hooks/use-upload.ts`)

#### `useCreateUploadSession()`

Mutation hook for creating upload sessions.

#### `useUploadFile()`

Mutation hook for uploading files.

#### `useUploadProgress(sessionId: string)`

Query hook for tracking upload progress with automatic polling.

#### `useFinalizeUpload()`

Mutation hook for finalizing uploads.

## Query Keys

Query keys are organized hierarchically for efficient cache management:

```typescript
// Auth keys
authKeys.all = ["auth"];
authKeys.session = ["auth", "session"];
authKeys.user = ["auth", "user"];
authKeys.admin = ["auth", "admin"];

// Audiobook keys
audiobookKeys.all = ["audiobooks"];
audiobookKeys.lists = ["audiobooks", "list"];
audiobookKeys.detail = (id) => ["audiobooks", "detail", id];

// Upload keys
uploadKeys.all = ["uploads"];
uploadKeys.detail = (id) => ["uploads", "detail", id];
uploadKeys.progress = (id) => ["uploads", "detail", id, "progress"];
```

## Usage Examples

### Basic Query

```typescript
import { useAudioBooks } from "@/hooks/use-audiobooks";

function AudioBooksList() {
  const { data, isLoading, error } = useAudioBooks();

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <div>
      {data?.data?.map((book) => (
        <div key={book.id}>{book.title}</div>
      ))}
    </div>
  );
}
```

### Mutation with Optimistic Updates

```typescript
import { useCreateAudioBook } from "@/hooks/use-audiobooks";

function CreateAudioBookForm() {
  const createMutation = useCreateAudioBook();

  const handleSubmit = (formData) => {
    createMutation.mutate(formData, {
      onSuccess: () => {
        // Handle success
      },
      onError: (error) => {
        // Handle error
      },
    });
  };

  return (
    <form onSubmit={handleSubmit}>
      {/* form fields */}
      <button disabled={createMutation.isPending}>
        {createMutation.isPending ? "Creating..." : "Create"}
      </button>
    </form>
  );
}
```

### Real-time Updates with Polling

```typescript
import { useUploadProgress } from "@/hooks/use-upload";

function UploadProgress({ sessionId }) {
  const { data: progress } = useUploadProgress(sessionId);

  return (
    <div>
      <div>Status: {progress?.status}</div>
      <div>
        Progress: {progress?.uploaded_files}/{progress?.total_files}
      </div>
    </div>
  );
}
```

## Cache Management

### Invalidation

Queries are automatically invalidated when related mutations succeed:

```typescript
// In useCreateAudioBook hook
onSuccess: () => {
  queryClient.invalidateQueries({ queryKey: audiobookKeys.lists() });
};
```

### Manual Cache Updates

For optimistic updates, you can manually update the cache:

```typescript
// In useUpdateAudioBook hook
onSuccess: (data, variables) => {
  queryClient.setQueryData(audiobookKeys.detail(variables.id), data);
};
```

## Error Handling

React Query provides built-in error handling with retry logic. Custom error handling can be added in mutation hooks:

```typescript
const mutation = useMutation({
  mutationFn: apiCall,
  onError: (error) => {
    console.error("Mutation failed:", error);
    // Show toast notification, etc.
  },
});
```

## Development Tools

The React Query DevTools are included in development mode and can be accessed by clicking the floating button in the bottom-right corner of the screen.

## Best Practices

1. **Use TypeScript**: All hooks are fully typed for better development experience.
2. **Organize Query Keys**: Use consistent naming and structure for query keys.
3. **Handle Loading States**: Always provide loading indicators for better UX.
4. **Error Boundaries**: Implement error boundaries for graceful error handling.
5. **Optimistic Updates**: Use optimistic updates for better perceived performance.
6. **Cache Invalidation**: Properly invalidate related queries after mutations.

## Migration from Server Components

The integration converts server-side data fetching to client-side queries, providing:

- Better user experience with loading states
- Automatic background updates
- Offline support capabilities
- Real-time data synchronization
