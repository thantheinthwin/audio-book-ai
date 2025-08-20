import { useMutation, useQueryClient } from "@tanstack/react-query";
import { uploadsAPI } from "@/lib/api";

// Query keys
export const uploadKeys = {
  all: ["uploads"] as const,
  lists: () => [...uploadKeys.all, "list"] as const,
  details: () => [...uploadKeys.all, "detail"] as const,
  detail: (id: string) => [...uploadKeys.details(), id] as const,
  progress: (id: string) => [...uploadKeys.detail(id), "progress"] as const,
};

// Hook to create upload session
export function useCreateUploadSession() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: uploadsAPI.createUploadSession,
    onSuccess: (data) => {
      // Add the new upload session to the cache
      queryClient.setQueryData(uploadKeys.detail(data.id), data);
      // Invalidate uploads list
      queryClient.invalidateQueries({ queryKey: uploadKeys.lists() });
    },
  });
}

// Hook to upload file
export function useUploadFile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      sessionId,
      file,
      chapterNumber,
      chapterTitle,
    }: {
      sessionId: string;
      file: File;
      chapterNumber?: number;
      chapterTitle?: string;
    }) => uploadsAPI.uploadFile(sessionId, file, chapterNumber, chapterTitle),
    onSuccess: (data, variables) => {
      // Update the upload session in cache
      queryClient.invalidateQueries({
        queryKey: uploadKeys.detail(variables.sessionId),
      });
    },
  });
}

// Hook to get upload progress
export function useUploadProgress(sessionId: string) {
  return useQuery({
    queryKey: uploadKeys.progress(sessionId),
    queryFn: () => uploadsAPI.getUploadProgress(sessionId),
    enabled: !!sessionId,
    refetchInterval: (data) => {
      // Stop polling if upload is complete
      if (data?.status === "completed" || data?.status === "failed") {
        return false;
      }
      return 2000; // Poll every 2 seconds
    },
  });
}

// Hook to finalize upload
export function useFinalizeUpload() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      sessionId,
      audiobookData,
    }: {
      sessionId: string;
      audiobookData: any;
    }) => uploadsAPI.finalizeUpload(sessionId, audiobookData),
    onSuccess: (data, variables) => {
      // Remove upload session from cache
      queryClient.removeQueries({
        queryKey: uploadKeys.detail(variables.sessionId),
      });
      // Invalidate uploads list
      queryClient.invalidateQueries({ queryKey: uploadKeys.lists() });
    },
  });
}
