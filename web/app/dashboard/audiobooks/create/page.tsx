"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Upload, X, FileAudio, Image as ImageIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { audiobooksAPI } from "@/lib/api";

export default function CreateAudioBookPage() {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const [audioFile, setAudioFile] = useState<File | null>(null);
  const [coverImage, setCoverImage] = useState<File | null>(null);
  const [formData, setFormData] = useState({
    title: "",
    author: "",
    description: "",
  });

  const handleFileUpload = (
    event: React.ChangeEvent<HTMLInputElement>,
    type: "audio" | "image"
  ) => {
    const file = event.target.files?.[0];
    if (file) {
      if (type === "audio") {
        setAudioFile(file);
      } else {
        setCoverImage(file);
      }
    }
  };

  const removeFile = (type: "audio" | "image") => {
    if (type === "audio") {
      setAudioFile(null);
    } else {
      setCoverImage(null);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      // TODO: Implement file upload to storage service
      // For now, we'll just create the audio book record
      const audioBookData = {
        title: formData.title,
        author: formData.author,
        description: formData.description,
        file_url: audioFile ? URL.createObjectURL(audioFile) : undefined,
        cover_image: coverImage ? URL.createObjectURL(coverImage) : undefined,
      };

      await audiobooksAPI.createAudioBook(audioBookData);
      router.push("/dashboard/audiobooks");
    } catch (error) {
      console.error("Failed to create audio book:", error);
      // TODO: Add proper error handling
    } finally {
      setIsLoading(false);
    }
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="font-bold text-3xl mb-2">Upload New Audio Book</h1>
        <p className="text-muted-foreground">
          Add a new audio book to your collection
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Basic Information */}
          <Card>
            <CardHeader>
              <CardTitle>Basic Information</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <Label htmlFor="title">Title *</Label>
                <Input
                  id="title"
                  value={formData.title}
                  onChange={(e) =>
                    setFormData({ ...formData, title: e.target.value })
                  }
                  placeholder="Enter audio book title"
                  required
                />
              </div>

              <div>
                <Label htmlFor="author">Author *</Label>
                <Input
                  id="author"
                  value={formData.author}
                  onChange={(e) =>
                    setFormData({ ...formData, author: e.target.value })
                  }
                  placeholder="Enter author name"
                  required
                />
              </div>

              <div>
                <Label htmlFor="description">Description</Label>
                <Textarea
                  id="description"
                  value={formData.description}
                  onChange={(e) =>
                    setFormData({ ...formData, description: e.target.value })
                  }
                  placeholder="Enter a brief description of the audio book"
                  rows={4}
                />
              </div>
            </CardContent>
          </Card>

          {/* File Upload */}
          <div className="space-y-6">
            {/* Audio File Upload */}
            <Card>
              <CardHeader>
                <CardTitle>Audio File *</CardTitle>
              </CardHeader>
              <CardContent>
                {!audioFile ? (
                  <div className="border-2 border-dashed border-muted-foreground/25 rounded-lg p-6 text-center">
                    <FileAudio className="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
                    <p className="text-sm text-muted-foreground mb-2">
                      Upload your audio file
                    </p>
                    <p className="text-xs text-muted-foreground mb-4">
                      Supported formats: MP3, WAV, M4A (Max 500MB)
                    </p>
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() =>
                        document.getElementById("audio-upload")?.click()
                      }
                    >
                      <Upload className="h-4 w-4 mr-2" />
                      Choose Audio File
                    </Button>
                    <input
                      id="audio-upload"
                      type="file"
                      accept="audio/*"
                      onChange={(e) => handleFileUpload(e, "audio")}
                      className="hidden"
                    />
                  </div>
                ) : (
                  <div className="border rounded-lg p-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <FileAudio className="h-5 w-5 text-blue-500" />
                        <div>
                          <p className="font-medium text-sm">
                            {audioFile.name}
                          </p>
                          <p className="text-xs text-muted-foreground">
                            {formatFileSize(audioFile.size)}
                          </p>
                        </div>
                      </div>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => removeFile("audio")}
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Cover Image Upload */}
            <Card>
              <CardHeader>
                <CardTitle>Cover Image</CardTitle>
              </CardHeader>
              <CardContent>
                {!coverImage ? (
                  <div className="border-2 border-dashed border-muted-foreground/25 rounded-lg p-6 text-center">
                    <ImageIcon className="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
                    <p className="text-sm text-muted-foreground mb-2">
                      Upload cover image
                    </p>
                    <p className="text-xs text-muted-foreground mb-4">
                      Recommended: 400x600px, JPG/PNG (Max 5MB)
                    </p>
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() =>
                        document.getElementById("image-upload")?.click()
                      }
                    >
                      <Upload className="h-4 w-4 mr-2" />
                      Choose Image
                    </Button>
                    <input
                      id="image-upload"
                      type="file"
                      accept="image/*"
                      onChange={(e) => handleFileUpload(e, "image")}
                      className="hidden"
                    />
                  </div>
                ) : (
                  <div className="border rounded-lg p-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <ImageIcon className="h-5 w-5 text-green-500" />
                        <div>
                          <p className="font-medium text-sm">
                            {coverImage.name}
                          </p>
                          <p className="text-xs text-muted-foreground">
                            {formatFileSize(coverImage.size)}
                          </p>
                        </div>
                      </div>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => removeFile("image")}
                      >
                        <X className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </div>

        {/* Submit Buttons */}
        <div className="flex gap-4 justify-end">
          <Button
            type="button"
            variant="outline"
            onClick={() => router.back()}
            disabled={isLoading}
          >
            Cancel
          </Button>
          <Button type="submit" disabled={isLoading || !audioFile}>
            {isLoading ? "Uploading..." : "Upload Audio Book"}
          </Button>
        </div>
      </form>
    </div>
  );
}
