// Query keys for uploads
export const uploadKeys = {
  all: ["uploads"] as const,
  lists: () => [...uploadKeys.all, "list"] as const,
  details: () => [...uploadKeys.all, "detail"] as const,
  detail: (id: string) => [...uploadKeys.details(), id] as const,
  progress: (id: string) => [...uploadKeys.detail(id), "progress"] as const,
};
