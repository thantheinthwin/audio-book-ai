import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { audiobooksAPI } from "@/lib/api";

// Query keys
export const audiobookKeys = {
  all: ["audiobooks"] as const,
  lists: () => [...audiobookKeys.all, "list"] as const,
  list: (filters: string) => [...audiobookKeys.lists(), { filters }] as const,
  details: () => [...audiobookKeys.all, "detail"] as const,
  detail: (id: string) => [...audiobookKeys.details(), id] as const,
};

// Hook to get all audiobooks
export function useAudioBooks() {
  return useQuery({
    queryKey: audiobookKeys.lists(),
    queryFn: () => audiobooksAPI.getAudioBooks(),
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}

// Hook to get a specific audiobook
export function useAudioBook(id: string) {
  return useQuery({
    queryKey: audiobookKeys.detail(id),
    queryFn: () => audiobooksAPI.getAudioBook(id),
    enabled: !!id,
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}

// Hook to create a new audiobook
export function useCreateAudioBook() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: audiobooksAPI.createAudioBook,
    onSuccess: () => {
      // Invalidate and refetch audiobooks list
      queryClient.invalidateQueries({ queryKey: audiobookKeys.lists() });
    },
  });
}

// Hook to update an audiobook
export function useUpdateAudioBook() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: any }) =>
      audiobooksAPI.updateAudioBook(id, data),
    onSuccess: (data, variables) => {
      // Update the specific audiobook in cache
      queryClient.setQueryData(audiobookKeys.detail(variables.id), data);
      // Invalidate and refetch audiobooks list
      queryClient.invalidateQueries({ queryKey: audiobookKeys.lists() });
    },
  });
}

// Hook to delete an audiobook
export function useDeleteAudioBook() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: audiobooksAPI.deleteAudioBook,
    onSuccess: (data, variables) => {
      // Remove the specific audiobook from cache
      queryClient.removeQueries({ queryKey: audiobookKeys.detail(variables) });
      // Invalidate and refetch audiobooks list
      queryClient.invalidateQueries({ queryKey: audiobookKeys.lists() });
    },
  });
}

// Hook to get audiobook chapters
export function useAudioBookChapters(audiobookId: string) {
  return useQuery({
    queryKey: [...audiobookKeys.detail(audiobookId), "chapters"],
    queryFn: () => audiobooksAPI.getAudioBookChapters(audiobookId),
    enabled: !!audiobookId,
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}
