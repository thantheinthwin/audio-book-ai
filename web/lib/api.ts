import { createClient } from "@/lib/supabase/client";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

// Helper function to get the current session token
async function getAuthToken(): Promise<string | null> {
  const supabase = createClient();
  const {
    data: { session },
  } = await supabase.auth.getSession();
  return session?.access_token || null;
}

// Generic API client with authentication
export async function apiClient<T = any>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const token = await getAuthToken();

  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...options.headers,
  };

  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    const error = await response
      .json()
      .catch(() => ({ error: "Unknown error" }));
    throw new Error(error.error || `HTTP error! status: ${response.status}`);
  }

  return response.json();
}

// Auth API functions
export const authAPI = {
  // Validate token
  validateToken: () => apiClient("/auth/validate", { method: "POST" }),

  // Get current user
  getMe: () => apiClient("/auth/me"),

  // Health check
  health: () => apiClient("/auth/health"),
};

// Profile API functions
export const profileAPI = {
  getProfile: () => apiClient("/profile"),

  updateProfile: (data: any) =>
    apiClient("/profile", {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  deleteProfile: () => apiClient("/profile", { method: "DELETE" }),
};

// Audio books API functions
export const audiobooksAPI = {
  getAudioBooks: () => apiClient("/audiobooks"),

  createAudioBook: (data: any) =>
    apiClient("/audiobooks", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  getAudioBook: (id: string) => apiClient(`/audiobooks/${id}`),

  updateAudioBook: (id: string, data: any) =>
    apiClient(`/audiobooks/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  deleteAudioBook: (id: string) =>
    apiClient(`/audiobooks/${id}`, {
      method: "DELETE",
    }),
};

// Public audio books API functions
export const publicAPI = {
  getPublicAudioBooks: () => apiClient("/public/audiobooks"),

  getPublicAudioBook: (id: string) => apiClient(`/public/audiobooks/${id}`),
};

// Library API functions
export const libraryAPI = {
  getLibrary: () => apiClient("/library"),

  addToLibrary: (audiobookId: string) =>
    apiClient(`/library/${audiobookId}`, {
      method: "POST",
    }),

  removeFromLibrary: (audiobookId: string) =>
    apiClient(`/library/${audiobookId}`, {
      method: "DELETE",
    }),
};

// Playlists API functions
export const playlistsAPI = {
  getPlaylists: () => apiClient("/playlists"),

  createPlaylist: (data: any) =>
    apiClient("/playlists", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  getPlaylist: (id: string) => apiClient(`/playlists/${id}`),

  updatePlaylist: (id: string, data: any) =>
    apiClient(`/playlists/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  deletePlaylist: (id: string) =>
    apiClient(`/playlists/${id}`, {
      method: "DELETE",
    }),

  addToPlaylist: (id: string, audiobookId: string) =>
    apiClient(`/playlists/${id}/items`, {
      method: "POST",
      body: JSON.stringify({ audiobookId }),
    }),

  removeFromPlaylist: (id: string, audiobookId: string) =>
    apiClient(`/playlists/${id}/items/${audiobookId}`, {
      method: "DELETE",
    }),
};

// Progress API functions
export const progressAPI = {
  getProgress: (audiobookId: string) => apiClient(`/progress/${audiobookId}`),

  updateProgress: (audiobookId: string, data: any) =>
    apiClient(`/progress/${audiobookId}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),
};

// Bookmarks API functions
export const bookmarksAPI = {
  getBookmarks: (audiobookId: string) => apiClient(`/bookmarks/${audiobookId}`),

  createBookmark: (audiobookId: string, data: any) =>
    apiClient(`/bookmarks/${audiobookId}`, {
      method: "POST",
      body: JSON.stringify(data),
    }),

  updateBookmark: (id: string, data: any) =>
    apiClient(`/bookmarks/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),

  deleteBookmark: (id: string) =>
    apiClient(`/bookmarks/${id}`, {
      method: "DELETE",
    }),
};
