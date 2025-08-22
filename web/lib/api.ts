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
  tags?: string[];
  duration_seconds?: number;
  file_size_bytes?: number;
  file_path: string;
  file_url?: string;
  cover_image_url?: string;
  language: string;
  is_public: boolean;
  price: number;
  status: string;
  created_by: string;
  created_at: string;
  updated_at: string;
  chapters?: Chapter[];
  // Frontend convenience fields
  cover_image?: string; // Alias for cover_image_url
  duration?: number; // Alias for duration_seconds
}

interface Chapter {
  id: string;
  audiobook_id: string;
  chapter_number: number;
  title: string;
  file_path: string;
  file_url?: string;
  file_size_bytes?: number;
  mime_type?: string;
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
  price?: number;
  cover_image_url?: string;
}

interface CreateAudioBookData {
  title: string;
  author: string;
  language: string;
  isPublic: boolean;
  price: number;
  coverImage?: File | null;
  chapters: Array<{
    id: string;
    chapter_number: number;
    title: string;
    audio_file?: File;
    playtime: number;
  }>;
}

// Cart types
interface CartItem {
  id: string;
  user_id: string;
  audiobook_id: string;
  added_at: string;
  audiobook: AudioBook;
}

interface CartResponse {
  items: CartItem[];
  total_items: number;
  total_price: number;
}

interface AddToCartRequest {
  audiobook_id: string;
}

export interface AddToCartResponse {
  cart_item_id: string;
  audiobook_id: string;
}

interface CartCheckResponse {
  is_in_cart: boolean;
}

// Checkout and Purchase types
interface CheckoutResponse {
  order_id: string;
  purchased_items: Array<{
    id: string;
    user_id: string;
    audiobook_id: string;
    purchase_price: number;
    purchased_at: string;
    transaction_id?: string;
    payment_status: string;
    audiobook: AudioBook;
  }>;
  total_amount: number;
  transaction_id: string;
  checkout_completed_at: string;
}

interface PurchaseHistoryResponse {
  purchases: Array<{
    id: string;
    user_id: string;
    audiobook_id: string;
    purchase_price: number;
    purchased_at: string;
    transaction_id?: string;
    payment_status: string;
    audiobook: AudioBook;
  }>;
  total_items: number;
  total_spent: number;
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
  getProfile: () => apiClient<ApiResponse<User>>("/user/profile"),

  updateProfile: (data: ProfileUpdateData) =>
    apiClient<ApiResponse<User>>("/user/profile", {
      method: "PUT",
      data,
    }),

  deleteProfile: () =>
    apiClient<ApiResponse>("/user/profile", { method: "DELETE" }),
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
  getAudioBooks: () => apiClient<ApiResponse<AudioBook[]>>("/user/audiobooks"),

  getAudioBook: (id: string) =>
    apiClient<ApiResponse<AudioBook>>(`/user/audiobooks/${id}`),

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
  createAudioBookWithFiles: async (data: CreateAudioBookData) => {
    try {
      const formData = new FormData();
      formData.append("title", data.title);
      formData.append("author", data.author);
      formData.append("language", data.language);
      formData.append("isPublic", data.isPublic.toString());
      formData.append("price", data.price.toString());

      if (data.coverImage) {
        formData.append("coverImage", data.coverImage);
      }

      // Add chapters metadata
      const chaptersMetadata = data.chapters.map((chapter) => ({
        id: chapter.id,
        chapter_number: chapter.chapter_number,
        title: chapter.title,
      }));
      formData.append("chapters", JSON.stringify(chaptersMetadata));

      // Add each file separately
      data.chapters.forEach((chapter, index) => {
        if (chapter.audio_file) {
          formData.append(`file_${index}`, chapter.audio_file);
          formData.append(
            `file_${index}_chapter_number`,
            chapter.chapter_number.toString()
          );
          formData.append(
            `file_${index}_title`,
            chapter.title || `Chapter ${chapter.chapter_number}`
          );
          formData.append(
            `file_${index}_duration_seconds`,
            chapter.playtime.toString()
          );
        }
      });

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

  updateAudioBookPrice: (id: string, price: number) =>
    apiClient<ApiResponse<AudioBook>>(`/admin/audiobooks/${id}/price`, {
      method: "PUT",
      data: { price },
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
          chapter_id?: string;
        }>;
        overall_status: string;
        progress: number;
        total_jobs: number;
        completed_jobs: number;
        failed_jobs: number;
      }>
    >(`/admin/audiobooks/${id}/jobs`),
};

// Cart API functions
export const cartAPI = {
  addToCart: (data: AddToCartRequest) =>
    apiClient<ApiResponse<AddToCartResponse>>("/user/cart", {
      method: "POST",
      data,
    }),

  removeFromCart: (audiobookId: string) =>
    apiClient<ApiResponse>(`/user/cart/${audiobookId}`, {
      method: "DELETE",
    }),

  getCart: () => apiClient<ApiResponse<CartResponse>>("/user/cart"),

  checkIfInCart: (audiobookId: string) =>
    apiClient<ApiResponse<CartCheckResponse>>(
      `/user/cart/${audiobookId}/check`
    ),
};

// Checkout and Purchase API functions
export const checkoutAPI = {
  checkout: (data: { cart_item_ids: string[] }) =>
    apiClient<ApiResponse<CheckoutResponse>>("/user/checkout", {
      method: "POST",
      data,
    }),

  getPurchaseHistory: (limit?: number, offset?: number) => {
    const params = new URLSearchParams();
    if (limit) params.append("limit", limit.toString());
    if (offset) params.append("offset", offset.toString());
    
    return apiClient<ApiResponse<PurchaseHistoryResponse>>(
      `/user/purchases?${params.toString()}`
    );
  },

  isAudioBookPurchased: (audiobookId: string) =>
    apiClient<ApiResponse<{ is_purchased: boolean }>>(
      `/user/audiobooks/${audiobookId}/purchased`
    ),
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
  CreateAudioBookData,
  UploadSession,
  UploadedFile,
  CartItem,
  CartResponse,
  AddToCartRequest,
  CartCheckResponse,
  CheckoutResponse,
  PurchaseHistoryResponse,
};
