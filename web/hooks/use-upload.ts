import { useMutation, useQueryClient, useQuery } from "@tanstack/react-query";
import { uploadAPI } from "@/lib/api";
import { uploadKeys } from "@/queryKeys";

// Hook to create upload session
export function useCreateUploadSession() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: uploadAPI.createUpload,
    onSuccess: (data) => {
      // Add the new upload session to the cache
      queryClient.setQueryData(
        uploadKeys.detail(data.data?.upload_id || ""),
        data
      );
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
    }) =>
      uploadAPI.uploadFile(sessionId, file, {
        chapter_number: chapterNumber,
        chapter_title: chapterTitle,
      }),
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
    queryFn: () => uploadAPI.getUploadProgress(sessionId),
    enabled: !!sessionId,
    refetchInterval: 2000, // Poll every 2 seconds
  });
}

// Hook to finalize upload
export function useFinalizeUpload() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ sessionId }: { sessionId: string }) =>
      uploadAPI.deleteUpload(sessionId),
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
