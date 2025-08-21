import { createClient } from "@/lib/supabase/client";
import axios, { AxiosInstance, AxiosResponse, AxiosError } from "axios";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

// Types for API responses
interface ApiResponse<T = unknown> {
  data?: T;
  message?: string;
  error?: string;
}

interface User {
  id: string;
  email: string;
  role: string;
  username?: string;
  first_name?: string;
  last_name?: string;
  is_active: boolean;
  is_verified: boolean;
  created_at: string;
  updated_at: string;
}

interface AudioBook {
  id: string;
  title: string;
  author: string;
  summary?: string;
  duration_seconds?: number;
  file_size_bytes?: number;
  file_path: string;
  file_url?: string;
  cover_image_url?: string;
  language: string;
  is_public: boolean;
  status: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

interface Chapter {
  id: string;
  audiobook_id: string;
  chapter_number: number;
  title: string;
  start_time_seconds?: number;
  end_time_seconds?: number;
  duration_seconds?: number;
  created_at: string;
}

// Note: AI processing types are not implemented in the current backend
// They would be added when the AI processing features are implemented

interface ProfileUpdateData {
  username?: string;
  first_name?: string;
  last_name?: string;
}

// Upload and Audio Book creation types
interface UploadSession {
  id: string;
  upload_type: "single" | "chapters";
  status: string;
  total_files: number;
  uploaded_files: number;
  total_size_bytes: number;
  created_at: string;
  updated_at: string;
}

interface UploadedFile {
  id: string;
  file_name: string;
  file_size_bytes: number;
  mime_type: string;
  chapter_number?: number;
  chapter_title?: string;
  status: string;
  uploaded_at: string;
}

interface AudioBookUpdateData {
  title?: string;
  author?: string;
  language?: string;
  is_public?: boolean;
  cover_image_url?: string;
}

// Helper function to get the current session token
async function getAuthToken(): Promise<string | null> {
  const supabase = createClient();
  const {
    data: { session },
  } = await supabase.auth.getSession();

  if (session?.access_token) {
    // Log the token for debugging (remove in production)
    console.log(
      "Auth token found:",
      session.access_token.substring(0, 20) + "..."
    );
    console.log("Full token length:", session.access_token.length);
    console.log("Token type:", typeof session.access_token);

    // Also log the user info
    console.log("User info:", {
      id: session.user?.id,
      email: session.user?.email,
      role: session.user?.user_metadata?.role || "user",
    });

    // Try to decode the JWT to see its structure
    try {
      const tokenParts = session.access_token.split(".");
      if (tokenParts.length === 3) {
        const payload = JSON.parse(atob(tokenParts[1]));
        console.log("JWT Token payload:", payload);
        console.log("JWT Token issuer:", payload.iss);
        console.log("JWT Token audience:", payload.aud);
        console.log("JWT Token subject:", payload.sub);
      }
    } catch (e) {
      console.log("Could not decode JWT token:", e);
    }

    return session.access_token;
  }

  console.log("No auth token found");
  return null;
}

// Create axios instance with default configuration
const createApiClient = (): AxiosInstance => {
  const client = axios.create({
    baseURL: API_BASE_URL,
    headers: {
      "Content-Type": "application/json",
    },
    timeout: 30000, // 30 seconds timeout
  });

  // Request interceptor to add auth token
  client.interceptors.request.use(
    async (config) => {
      const token = await getAuthToken();
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
        console.log("Request with auth token:", config.url);
      } else {
        console.log("Request without auth token:", config.url);
      }
      return config;
    },
    (error) => {
      return Promise.reject(error);
    }
  );

  // Response interceptor for error handling
  client.interceptors.response.use(
    (response: AxiosResponse) => {
      return response;
    },
    (error: AxiosError) => {
      if (error.response) {
        // Server responded with error status
        const errorData = error.response.data as any;
        console.error("API Error Response:", {
          status: error.response.status,
          statusText: error.response.statusText,
          data: errorData,
          headers: error.response.headers,
        });

        const errorMessage =
          errorData?.error ||
          errorData?.message ||
          `HTTP error! status: ${error.response.status}`;
        return Promise.reject(new Error(errorMessage));
      } else if (error.request) {
        // Request was made but no response received
        console.error("API Error: No response received", error.request);
        return Promise.reject(new Error("No response received from server"));
      } else {
        // Something else happened
        console.error("API Error: Request failed", error.message);
        return Promise.reject(new Error("Request failed"));
      }
    }
  );

  return client;
};

// Generic API client with authentication
export async function apiClient<T = unknown>(
  endpoint: string,
  options: {
    method?: "GET" | "POST" | "PUT" | "DELETE" | "PATCH";
    data?: any;
    params?: any;
    headers?: Record<string, string>;
  } = {}
): Promise<T> {
  const client = createApiClient();
  const { method = "GET", data, params, headers } = options;

  try {
    const response = await client.request({
      url: `/api/v1${endpoint}`,
      method,
      data,
      params,
      headers,
    });

    return response.data;
  } catch (error) {
    throw error;
  }
}

// Auth API functions
export const authAPI = {
  // Validate token
  validateToken: () =>
    apiClient<ApiResponse>("/auth/validate", { method: "POST" }),

  // Get current user
  getMe: () => apiClient<ApiResponse<User>>("/auth/me"),

  // Health check
  health: () => apiClient<ApiResponse>("/auth/health"),
};

// Profile API functions
export const profileAPI = {
  getProfile: () => apiClient<ApiResponse<User>>("/profile"),

  updateProfile: (data: ProfileUpdateData) =>
    apiClient<ApiResponse<User>>("/profile", {
      method: "PUT",
      data,
    }),

  deleteProfile: () => apiClient<ApiResponse>("/profile", { method: "DELETE" }),
};

// Upload API functions
export const uploadAPI = {
  // Create upload session
  createUpload: (data: {
    upload_type: "single" | "chapters";
    total_files: number;
    total_size_bytes: number;
  }) =>
    apiClient<
      ApiResponse<{ upload_id: string; status: string; message: string }>
    >("/admin/uploads", {
      method: "POST",
      data,
    }),

  // Upload file to session
  uploadFile: (
    uploadId: string,
    file: File,
    metadata?: {
      chapter_number?: number;
      chapter_title?: string;
    }
  ) => {
    const formData = new FormData();
    formData.append("file", file);

    if (metadata?.chapter_number) {
      formData.append("chapter_number", metadata.chapter_number.toString());
    }
    if (metadata?.chapter_title) {
      formData.append("chapter_title", metadata.chapter_title);
    }

    return apiClient<
      ApiResponse<{
        file_id: string;
        upload_id: string;
        file_name: string;
        file_size_bytes: number;
        uploaded_at: string;
        chapter_number?: number;
        chapter_title?: string;
      }>
    >(`/admin/uploads/${uploadId}/files`, {
      method: "POST",
      data: formData,
      headers: {
        "Content-Type": "multipart/form-data",
      },
    });
  },

  // Get upload progress
  getUploadProgress: (uploadId: string) =>
    apiClient<
      ApiResponse<{
        upload_id: string;
        status: string;
        total_files: number;
        uploaded_files: number;
        failed_files: number;
        retrying_files: number;
        progress: number;
        total_size_bytes: number;
        uploaded_size_bytes: number;
        files?: Array<{
          id: string;
          file_name: string;
          file_size_bytes: number;
          mime_type: string;
          chapter_number?: number;
          chapter_title?: string;
          status: string;
          error?: string;
          uploaded_at: string;
        }>;
      }>
    >(`/admin/uploads/${uploadId}/progress`),

  // Upload files in batch
  uploadFilesBatch: (
    uploadId: string,
    files: File[],
    metadata: Array<{
      chapter_number?: number;
      chapter_title?: string;
    }>
  ) => {
    const formData = new FormData();

    files.forEach((file, index) => {
      formData.append("files", file);
      formData.append(
        "chapter_numbers",
        metadata[index]?.chapter_number?.toString() || ""
      );
      formData.append("chapter_titles", metadata[index]?.chapter_title || "");
    });

    return apiClient<
      ApiResponse<{
        upload_id: string;
        total_files: number;
        success_count: number;
        failed_count: number;
        retrying_count: number;
        files: Array<{
          file_id: string;
          upload_id: string;
          file_name: string;
          file_size_bytes: number;
          status: string;
          retry_count: number;
          uploaded_at: string;
          chapter_number?: number;
          chapter_title?: string;
        }>;
        errors?: string[];
      }>
    >(`/admin/uploads/${uploadId}/files/batch`, {
      method: "POST",
      data: formData,
      headers: {
        "Content-Type": "multipart/form-data",
      },
    });
  },

  // Retry failed upload
  retryFailedUpload: (uploadId: string, fileId: string) =>
    apiClient<
      ApiResponse<{
        message: string;
        file_id: string;
        retry_count: number;
      }>
    >(`/admin/uploads/${uploadId}/files/${fileId}/retry`, {
      method: "POST",
    }),

  // Get upload details
  getUploadDetails: (uploadId: string) =>
    apiClient<
      ApiResponse<{
        upload: {
          id: string;
          upload_type: string;
          status: string;
          total_files: number;
          uploaded_files: number;
          total_size_bytes: number;
          created_at: string;
          updated_at: string;
        };
        files: Array<{
          id: string;
          file_name: string;
          file_size_bytes: number;
          mime_type: string;
          chapter_number?: number;
          chapter_title?: string;
          status: string;
          uploaded_at: string;
        }>;
      }>
    >(`/admin/uploads/${uploadId}`),

  // Delete upload
  deleteUpload: (uploadId: string) =>
    apiClient<ApiResponse>(`/admin/uploads/${uploadId}`, {
      method: "DELETE",
    }),
};

// Audio books API functions
export const audiobooksAPI = {
  // User operations (protected routes)
  getAudioBooks: () => apiClient<ApiResponse<AudioBook[]>>("/audiobooks"),

  getAudioBook: (id: string) =>
    apiClient<ApiResponse<AudioBook>>(`/audiobooks/${id}`),

  // Admin operations (admin routes)
  createAudioBook: (data: {
    upload_id: string;
    title: string;
    author: string;
    language: string;
    is_public: boolean;
    cover_image_url?: string;
  }) =>
    apiClient<
      ApiResponse<{
        audiobook_id: string;
        status: string;
        message: string;
        jobs_created: number;
      }>
    >("/admin/audiobooks", {
      method: "POST",
      data,
    }),

  // New function to create audio book using Next.js API route
  createAudioBookWithFiles: async (formData: FormData) => {
    try {
      const response = await fetch("/api/audiobooks/create", {
        method: "POST",
        body: formData,
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || "Failed to create audio book");
      }

      return await response.json();
    } catch (error) {
      throw error;
    }
  },

  updateAudioBook: (id: string, data: AudioBookUpdateData) =>
    apiClient<ApiResponse<AudioBook>>(`/admin/audiobooks/${id}`, {
      method: "PUT",
      data,
    }),

  deleteAudioBook: (id: string) =>
    apiClient<ApiResponse>(`/admin/audiobooks/${id}`, {
      method: "DELETE",
    }),

  // Get job status for audio book
  getJobStatus: (id: string) =>
    apiClient<
      ApiResponse<{
        audiobook_id: string;
        jobs: Array<{
          id: string;
          job_type: string;
          status: string;
          created_at: string;
          started_at?: string;
          completed_at?: string;
          error_message?: string;
        }>;
        overall_status: string;
        progress: number;
        total_jobs: number;
        completed_jobs: number;
        failed_jobs: number;
      }>
    >(`/admin/audiobooks/${id}/jobs`),
};

// Note: AI Processing API functions are not implemented in the current backend
// They would be added when the AI processing features are implemented

// Public audio books API functions (no auth required)
export const publicAPI = {
  getPublicAudioBooks: () => apiClient<ApiResponse<AudioBook[]>>("/audiobooks"),

  getPublicAudioBook: (id: string) =>
    apiClient<ApiResponse<AudioBook>>(`/audiobooks/${id}`),
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
      data,
    }),

  getPlaylist: (id: string) => apiClient(`/playlists/${id}`),

  updatePlaylist: (id: string, data: any) =>
    apiClient(`/playlists/${id}`, {
      method: "PUT",
      data,
    }),

  deletePlaylist: (id: string) =>
    apiClient(`/playlists/${id}`, {
      method: "DELETE",
    }),

  addToPlaylist: (id: string, audiobookId: string) =>
    apiClient(`/playlists/${id}/items`, {
      method: "POST",
      data: { audiobookId },
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
      data,
    }),
};

// Bookmarks API functions
export const bookmarksAPI = {
  getBookmarks: (audiobookId: string) => apiClient(`/bookmarks/${audiobookId}`),

  createBookmark: (audiobookId: string, data: any) =>
    apiClient(`/bookmarks/${audiobookId}`, {
      method: "POST",
      data,
    }),

  updateBookmark: (id: string, data: any) =>
    apiClient(`/bookmarks/${id}`, {
      method: "PUT",
      data,
    }),

  deleteBookmark: (id: string) =>
    apiClient(`/bookmarks/${id}`, {
      method: "DELETE",
    }),
};

// Test API without authentication
export const testApiWithoutAuth = async () => {
  try {
    const client = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        "Content-Type": "application/json",
      },
      timeout: 30000,
    });

    console.log("Testing API without authentication...");
    const response = await client.get("/api/v1/auth/health");
    console.log("API test without auth successful:", response.data);
    return true;
  } catch (error) {
    console.error("API test without auth failed:", error);
    return false;
  }
};

// Test admin authentication
export const testAdminAuth = async () => {
  try {
    console.log("Testing admin authentication...");
    // const response = await apiClient<ApiResponse>("/audiobooks");
    const response = await audiobooksAPI.getAudioBooks();
    console.log("Admin auth test successful:", response);
    return true;
  } catch (error) {
    console.error("Admin auth test failed:", error);
    return false;
  }
};

// Test authentication function
export const testAuth = async () => {
  try {
    const supabase = createClient();
    const {
      data: { session },
    } = await supabase.auth.getSession();

    console.log("Current session:", {
      hasSession: !!session,
      hasAccessToken: !!session?.access_token,
      tokenLength: session?.access_token?.length || 0,
      user: session?.user?.email,
    });

    if (session?.access_token) {
      // Decode the JWT token to see its structure
      try {
        const tokenParts = session.access_token.split(".");
        if (tokenParts.length === 3) {
          const payload = JSON.parse(atob(tokenParts[1]));
          console.log("JWT Token payload:", payload);
        }
      } catch (e) {
        console.log("Could not decode JWT token:", e);
      }

      // Test the token with an authenticated endpoint
      const response = await apiClient<ApiResponse>("/auth/me");
      console.log("Auth test successful:", response);
      return true;
    } else {
      console.log("No session found");
      return false;
    }
  } catch (error) {
    console.error("Auth test failed:", error);
    return false;
  }
};

// Export types for use in components
export type {
  ApiResponse,
  User,
  AudioBook,
  Chapter,
  ProfileUpdateData,
  AudioBookUpdateData,
  UploadSession,
  UploadedFile,
};
