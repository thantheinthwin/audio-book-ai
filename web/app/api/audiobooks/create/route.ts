import { NextRequest, NextResponse } from "next/server";
import { createClient } from "@/lib/supabase/server";
import axios from "axios";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export async function POST(request: NextRequest) {
  try {
    // Get the Supabase client for server-side operations
    const supabase = await createClient();

    // Get the current session
    const {
      data: { session },
    } = await supabase.auth.getSession();

    if (!session?.access_token) {
      return NextResponse.json(
        { error: "Authentication required" },
        { status: 401 }
      );
    }

    // Parse the request body
    const formData = await request.formData();
    const title = formData.get("title") as string;
    const author = formData.get("author") as string;
    const language = formData.get("language") as string;
    const isPublic = formData.get("isPublic") === "true";
    const coverImage = formData.get("coverImage") as File | null;
    const chapters = JSON.parse(formData.get("chapters") as string);

    // Validate required fields
    if (!title || !author || !language) {
      return NextResponse.json(
        { error: "Title, author, and language are required" },
        { status: 400 }
      );
    }

    if (!chapters || chapters.length === 0) {
      return NextResponse.json(
        { error: "At least one chapter is required" },
        { status: 400 }
      );
    }

    const chaptersWithFiles = chapters.filter(
      (chapter: any) => chapter.audio_file
    );
    if (chaptersWithFiles.length === 0) {
      return NextResponse.json(
        { error: "At least one audio file is required" },
        { status: 400 }
      );
    }

    // Step 1: Create upload session
    const totalFiles = chaptersWithFiles.length;
    const totalSize = chaptersWithFiles.reduce((sum: number, chapter: any) => {
      return sum + (chapter.audio_file?.size || 0);
    }, 0);

    let uploadData;
    try {
      const uploadResponse = await axios.post(
        `${API_BASE_URL}/api/v1/admin/uploads`,
        {
          upload_type: totalFiles === 1 ? "single" : "chapters",
          total_files: totalFiles,
          total_size_bytes: totalSize,
        },
        {
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${session.access_token}`,
          },
        }
      );

      uploadData = uploadResponse.data;
    } catch (error: any) {
      console.error(
        "Upload session creation failed:",
        error.response?.data || error.message
      );
      return NextResponse.json(
        {
          error: "Failed to create upload session",
          details: error.response?.data || error.message,
        },
        { status: error.response?.status || 500 }
      );
    }
    const uploadId = uploadData.data?.upload_id;

    if (!uploadId) {
      return NextResponse.json(
        { error: "Failed to get upload ID from response" },
        { status: 500 }
      );
    }

    // Step 2: Upload each file
    const uploadedFiles = [];
    for (let i = 0; i < chaptersWithFiles.length; i++) {
      const chapter = chaptersWithFiles[i];
      if (!chapter.audio_file) continue;

      const fileFormData = new FormData();
      fileFormData.append("file", chapter.audio_file);
      fileFormData.append("chapter_number", chapter.chapter_number.toString());
      fileFormData.append(
        "chapter_title",
        chapter.title || `Chapter ${chapter.chapter_number}`
      );

      try {
        const fileUploadResponse = await axios.post(
          `${API_BASE_URL}/api/v1/admin/uploads/${uploadId}/files`,
          fileFormData,
          {
            headers: {
              Authorization: `Bearer ${session.access_token}`,
              "Content-Type": "multipart/form-data",
            },
          }
        );

        const fileData = fileUploadResponse.data;
        uploadedFiles.push(fileData);
      } catch (error: any) {
        console.error(
          `File upload failed for chapter ${i + 1}:`,
          error.response?.data || error.message
        );
        return NextResponse.json(
          {
            error: `Failed to upload file for chapter ${i + 1}`,
            details: error.response?.data || error.message,
          },
          { status: error.response?.status || 500 }
        );
      }
    }

    // Step 3: Create audio book from upload
    const audioBookData = {
      upload_id: uploadId,
      title,
      author,
      language,
      is_public: isPublic,
      cover_image_url: coverImage
        ? await uploadCoverImage(coverImage)
        : undefined,
    };

    let audioBookResult;
    try {
      const audioBookResponse = await axios.post(
        `${API_BASE_URL}/api/v1/admin/audiobooks`,
        audioBookData,
        {
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${session.access_token}`,
          },
        }
      );

      audioBookResult = audioBookResponse.data;
    } catch (error: any) {
      console.error(
        "Audio book creation failed:",
        error.response?.data || error.message
      );
      return NextResponse.json(
        {
          error: "Failed to create audio book",
          details: error.response?.data || error.message,
        },
        { status: error.response?.status || 500 }
      );
    }

    return NextResponse.json({
      success: true,
      data: audioBookResult.data,
      message: "Audio book created successfully",
    });
  } catch (error) {
    console.error("Audio book creation error:", error);
    return NextResponse.json(
      {
        error: "Internal server error",
        details: error instanceof Error ? error.message : "Unknown error",
      },
      { status: 500 }
    );
  }
}

// Helper function to upload cover image
async function uploadCoverImage(file: File): Promise<string> {
  // For now, we'll return a placeholder URL
  // In a real implementation, you would upload to your storage service
  // and return the actual URL
  return `https://placeholder.com/cover/${file.name}`;
}
