// Query keys for authentication
export const authKeys = {
  all: ["auth"] as const,
  session: () => [...authKeys.all, "session"] as const,
  user: () => [...authKeys.all, "user"] as const,
  admin: () => [...authKeys.all, "admin"] as const,
};
