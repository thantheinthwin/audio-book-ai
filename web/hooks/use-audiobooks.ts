import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { audiobooksAPI } from "@/lib/api";
import { audiobookKeys } from "@/queryKeys";

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

// Hook to get job status for an audiobook
export function useAudioBookJobStatus(id: string, isPublicUser: boolean) {
  return useQuery({
    queryKey: [...audiobookKeys.detail(id), "jobs"],
    queryFn: () => audiobooksAPI.getJobStatus(id),
    enabled: !!id && !isPublicUser,
    refetchInterval: (data) => {
      // If still processing, refetch every 5 seconds
      if ((data?.state?.data?.data as any)?.overall_status !== "completed") {
        return 5000;
      }
      // If completed or failed, stop refetching
      return false;
    },
    staleTime: 1000 * 30, // 30 seconds
  });
}

// Hook to create a new audiobook with files
export function useCreateAudioBookWithFiles() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: audiobooksAPI.createAudioBookWithFiles,
    onSuccess: () => {
      // Invalidate and refetch audiobooks list
      queryClient.invalidateQueries({ queryKey: audiobookKeys.lists() });
    },
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

// Hook to update audiobook price (admin only)
export function useUpdateAudioBookPrice() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, price }: { id: string; price: number }) =>
      audiobooksAPI.updateAudioBookPrice(id, price),
    onSuccess: (data, variables) => {
      queryClient.invalidateQueries({
        queryKey: audiobookKeys.detail(variables.id),
      });
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
// Note: This function is not implemented in the current backend
// export function useAudioBookChapters(audiobookId: string) {
//   return useQuery({
//     queryKey: [...audiobookKeys.detail(audiobookId), "chapters"],
//     queryFn: () => audiobooksAPI.getAudioBookChapters(audiobookId),
//     enabled: !!audiobookId,
//     staleTime: 1000 * 60 * 5, // 5 minutes
//   });
// }
