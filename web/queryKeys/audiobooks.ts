// Query keys for audiobooks
export const audiobookKeys = {
  all: ["audiobooks"] as const,
  lists: () => [...audiobookKeys.all, "list"] as const,
  list: (filters: string) => [...audiobookKeys.lists(), { filters }] as const,
  details: () => [...audiobookKeys.all, "detail"] as const,
  detail: (id: string) => [...audiobookKeys.details(), id] as const,
};
