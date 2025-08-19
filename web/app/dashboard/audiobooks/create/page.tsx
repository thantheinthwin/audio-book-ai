"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import { Progress } from "@/components/ui/progress";
import { audiobooksAPI } from "@/lib/api";
import { Upload, FileAudio, CheckCircle, AlertCircle } from "lucide-react";
import CoverImageUpload from "@/components/cover-image-upload";
import AudioFilesUpload from "@/components/audio-files-upload";

interface FormData {
  title: string;
  author: string;
  language: string;
  isPublic: boolean;
  chapters: Array<{
    id: string;
    chapter_number: number;
    title: string;
    audio_file?: File;
  }>;
}

interface UploadState {
  uploadId: string | null;
  status: "idle" | "creating" | "uploading" | "completed" | "error";
  progress: number;
  uploadedFiles: number;
  totalFiles: number;
  error: string | null;
}

export default function CreateAudioBookPage() {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const [coverImage, setCoverImage] = useState<File | null>(null);
  const [uploadState, setUploadState] = useState<UploadState>({
    uploadId: null,
    status: "idle",
    progress: 0,
    uploadedFiles: 0,
    totalFiles: 0,
    error: null,
  });

  const {
    register,
    control,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = useForm<FormData>({
    defaultValues: {
      title: "",
      author: "",
      language: "en",
      isPublic: false,
      chapters: [
        {
          id: "chapter-1",
          chapter_number: 1,
          title: "",
        },
      ],
    },
  });

  const onSubmit = async (data: FormData) => {
    if (data.chapters.length === 0) {
      alert("Please add at least one chapter");
      return;
    }

    const chaptersWithFiles = data.chapters.filter(
      (chapter) => chapter.audio_file
    );
    if (chaptersWithFiles.length === 0) {
      alert("Please upload at least one audio file");
      return;
    }

    setIsLoading(true);
    setUploadState((prev) => ({ ...prev, status: "creating", error: null }));

    try {
      // Create FormData for the Next.js API route
      const formData = new FormData();
      formData.append("title", data.title);
      formData.append("author", data.author);
      formData.append("language", data.language);
      formData.append("isPublic", data.isPublic.toString());

      if (coverImage) {
        formData.append("coverImage", coverImage);
      }

      formData.append("chapters", JSON.stringify(data.chapters));

      // Call the Next.js API route
      const response = await audiobooksAPI.createAudioBookWithFiles(formData);

      console.log("Audio book created:", response);
      router.push("/dashboard/audiobooks");
    } catch (error) {
      console.error("Failed to create audio book:", error);
      setUploadState((prev) => ({
        ...prev,
        status: "error",
        error: error instanceof Error ? error.message : "Unknown error",
      }));
    } finally {
      setIsLoading(false);
    }
  };

  const getStatusIcon = () => {
    switch (uploadState.status) {
      case "idle":
        return <Upload className="h-5 w-5" />;
      case "creating":
        return <FileAudio className="h-5 w-5 animate-pulse" />;
      case "uploading":
        return <FileAudio className="h-5 w-5 animate-pulse" />;
      case "completed":
        return <CheckCircle className="h-5 w-5 text-green-500" />;
      case "error":
        return <AlertCircle className="h-5 w-5 text-red-500" />;
      default:
        return <Upload className="h-5 w-5" />;
    }
  };

  const getStatusText = () => {
    switch (uploadState.status) {
      case "idle":
        return "Ready to upload";
      case "creating":
        return "Creating upload session...";
      case "uploading":
        return `Uploading files... (${uploadState.uploadedFiles}/${uploadState.totalFiles})`;
      case "completed":
        return "Upload completed! Creating audio book...";
      case "error":
        return `Error: ${uploadState.error}`;
      default:
        return "Ready to upload";
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Create New Audio Book</h1>
        <p className="text-muted-foreground">
          Upload your audio files and create a new audio book
        </p>
      </div>

      {/* Upload Progress */}
      {uploadState.status !== "idle" && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              {getStatusIcon()}
              Upload Progress
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between text-sm">
              <span>{getStatusText()}</span>
              {uploadState.status === "uploading" && (
                <span>{Math.round(uploadState.progress)}%</span>
              )}
            </div>
            {uploadState.status === "uploading" && (
              <Progress value={uploadState.progress} className="w-full" />
            )}
            {uploadState.error && (
              <div className="text-red-500 text-sm">{uploadState.error}</div>
            )}
          </CardContent>
        </Card>
      )}

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        {/* Basic Information */}
        <Card>
          <CardHeader>
            <CardTitle>Basic Information</CardTitle>
            <CardDescription>
              Enter the basic details about your audio book
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4 grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="flex flex-col gap-4">
              <div className="space-y-2">
                <Label htmlFor="title">Title *</Label>
                <Input
                  id="title"
                  {...register("title", { required: "Title is required" })}
                  placeholder="Enter audio book title"
                />
                {errors.title && (
                  <p className="text-sm text-red-500">{errors.title.message}</p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="author">Author *</Label>
                <Input
                  id="author"
                  {...register("author", { required: "Author is required" })}
                  placeholder="Enter author name"
                />
                {errors.author && (
                  <p className="text-sm text-red-500">
                    {errors.author.message}
                  </p>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="language">Language *</Label>
                <Select
                  value={watch("language")}
                  onValueChange={(value) => setValue("language", value)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select language" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="en">English</SelectItem>
                    <SelectItem value="es">Spanish</SelectItem>
                    <SelectItem value="fr">French</SelectItem>
                    <SelectItem value="de">German</SelectItem>
                    <SelectItem value="it">Italian</SelectItem>
                    <SelectItem value="pt">Portuguese</SelectItem>
                    <SelectItem value="ru">Russian</SelectItem>
                    <SelectItem value="ja">Japanese</SelectItem>
                    <SelectItem value="ko">Korean</SelectItem>
                    <SelectItem value="zh">Chinese</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="flex items-center space-x-2">
                <Checkbox
                  id="isPublic"
                  checked={watch("isPublic")}
                  onCheckedChange={(checked) =>
                    setValue("isPublic", checked as boolean)
                  }
                />
                <Label htmlFor="isPublic">Make this audio book public</Label>
              </div>
            </div>

            <div className="space-y-2">
              <Label>Cover Image (Optional)</Label>
              <CoverImageUpload
                maxSizeMB={5}
                onImageChange={setCoverImage}
                className="w-full"
              />
            </div>
          </CardContent>
        </Card>

        {/* Audio Files */}
        <Card>
          <CardHeader>
            <CardTitle>Audio Files</CardTitle>
            <CardDescription>
              Upload your audio files. You can upload a single file or multiple
              chapter files. Each file will become a chapter in your audio book.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <AudioFilesUpload
              maxSizeMB={100}
              maxFiles={20}
              control={control}
              register={register}
              name="chapters"
              className="w-full"
            />
          </CardContent>
        </Card>

        {/* Submit Button */}
        <div className="flex justify-end">
          <Button
            type="submit"
            disabled={isLoading || uploadState.status === "uploading"}
            className="min-w-[200px]"
          >
            {isLoading ? "Creating..." : "Create Audio Book"}
          </Button>
        </div>
      </form>
    </div>
  );
}
