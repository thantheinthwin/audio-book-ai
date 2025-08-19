import { redirect } from "next/navigation";
import { createClient } from "@/lib/supabase/server";
import { audiobooksAPI } from "@/lib/api";
import { notFound } from "next/navigation";
import { Play, Edit, Trash2, Download, Share2 } from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";

interface PageProps {
  params: {
    id: string;
  };
}

export default async function AudioBookDetailPage({ params }: PageProps) {
  const supabase = await createClient();

  const { data, error } = await supabase.auth.getClaims();
  if (error || !data?.claims) {
    redirect("/auth/login");
  }

  // Check if user is admin
  const userRole = data.claims.user_metadata.role || "user";
  if (userRole !== "admin") {
    redirect("/");
  }

  // Fetch audio book details
  let audioBook = null;

  try {
    const audioBookResponse = await audiobooksAPI.getAudioBook(params.id);
    audioBook = audioBookResponse.data;
  } catch (error) {
    console.error("Failed to fetch audio book details:", error);
  }

  if (!audioBook) {
    notFound();
  }

  const formatDuration = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    return `${hours}:${minutes.toString().padStart(2, "0")}:${secs
      .toString()
      .padStart(2, "0")}`;
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-start">
        <div>
          <h1 className="font-bold text-3xl mb-2">{audioBook.title}</h1>
          <p className="text-muted-foreground text-lg">by {audioBook.author}</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" asChild>
            <Link href={`/dashboard/audiobooks/${params.id}/edit`}>
              <Edit className="h-4 w-4 mr-2" />
              Edit
            </Link>
          </Button>
          <Button variant="outline" className="text-red-600 hover:text-red-700">
            <Trash2 className="h-4 w-4 mr-2" />
            Delete
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Cover Image and Basic Info */}
          <Card>
            <CardContent className="p-6">
              <div className="flex gap-6">
                {audioBook.cover_image && (
                  <img
                    src={audioBook.cover_image}
                    alt={audioBook.title}
                    className="w-32 h-48 object-cover rounded-md"
                  />
                )}
                <div className="flex-1">
                  <div className="space-y-4">
                    <div className="flex gap-4">
                      {audioBook.duration && (
                        <div>
                          <span className="text-sm text-muted-foreground">
                            Duration:
                          </span>
                          <p className="font-medium">
                            {formatDuration(audioBook.duration)}
                          </p>
                        </div>
                      )}
                      <div>
                        <span className="text-sm text-muted-foreground">
                          Added:
                        </span>
                        <p className="font-medium">
                          {new Date(audioBook.created_at).toLocaleDateString()}
                        </p>
                      </div>
                    </div>

                    <div className="flex gap-2">
                      <Button size="sm">
                        <Play className="h-4 w-4 mr-2" />
                        Play
                      </Button>
                      <Button size="sm" variant="outline">
                        <Download className="h-4 w-4 mr-2" />
                        Download
                      </Button>
                      <Button size="sm" variant="outline">
                        <Share2 className="h-4 w-4 mr-2" />
                        Share
                      </Button>
                    </div>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Audio Book Information */}
          <Card>
            <CardHeader>
              <CardTitle>Audio Book Information</CardTitle>
              <CardDescription>
                Basic information about this audio book
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <span className="text-sm text-muted-foreground">Title:</span>
                <p className="font-medium">{audioBook.title}</p>
              </div>
              <div>
                <span className="text-sm text-muted-foreground">Author:</span>
                <p className="font-medium">{audioBook.author}</p>
              </div>
              {audioBook.summary && (
                <div>
                  <span className="text-sm text-muted-foreground">
                    Summary:
                  </span>
                  <p className="text-sm mt-1">{audioBook.summary}</p>
                </div>
              )}
              <div>
                <span className="text-sm text-muted-foreground">Created:</span>
                <p className="text-sm">
                  {new Date(audioBook.created_at).toLocaleString()}
                </p>
              </div>
              <div>
                <span className="text-sm text-muted-foreground">
                  Last Updated:
                </span>
                <p className="text-sm">
                  {new Date(audioBook.updated_at).toLocaleString()}
                </p>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Quick Actions */}
          <Card>
            <CardHeader>
              <CardTitle>Quick Actions</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <Button className="w-full" size="sm">
                <Play className="h-4 w-4 mr-2" />
                Play Audio Book
              </Button>
              <Button className="w-full" variant="outline" size="sm">
                <Download className="h-4 w-4 mr-2" />
                Download
              </Button>
              <Button className="w-full" variant="outline" size="sm">
                <Share2 className="h-4 w-4 mr-2" />
                Share
              </Button>
            </CardContent>
          </Card>

          {/* File Information */}
          <Card>
            <CardHeader>
              <CardTitle>File Information</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div>
                <span className="text-sm text-muted-foreground">File URL:</span>
                <p className="text-sm font-mono break-all">
                  {audioBook.file_url || "Not available"}
                </p>
              </div>
              <Separator />
              <div>
                <span className="text-sm text-muted-foreground">
                  Cover Image:
                </span>
                <p className="text-sm font-mono break-all">
                  {audioBook.cover_image || "Not available"}
                </p>
              </div>
              <Separator />
              <div>
                <span className="text-sm text-muted-foreground">
                  Last Updated:
                </span>
                <p className="text-sm">
                  {new Date(audioBook.updated_at).toLocaleString()}
                </p>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
