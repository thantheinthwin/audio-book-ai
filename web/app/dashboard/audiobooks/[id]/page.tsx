import { redirect } from "next/navigation";
import { createClient } from "@/lib/supabase/server";
import { audiobooksAPI, aiProcessingAPI } from "@/lib/api";
import { notFound } from "next/navigation";
import {
  Play,
  Pause,
  SkipBack,
  SkipForward,
  Volume2,
  Edit,
  Trash2,
  FileText,
  Brain,
  Tags,
  Download,
  Share2,
} from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

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
  let transcript = null;
  let summary = null;
  let tags = null;
  let processingJobs = [];

  try {
    const [
      audioBookResponse,
      transcriptResponse,
      summaryResponse,
      tagsResponse,
      jobsResponse,
    ] = await Promise.allSettled([
      audiobooksAPI.getAudioBook(params.id),
      aiProcessingAPI.getTranscript(params.id),
      aiProcessingAPI.getSummary(params.id),
      aiProcessingAPI.getTags(params.id),
      aiProcessingAPI.getProcessingJobs(params.id),
    ]);

    if (audioBookResponse.status === "fulfilled") {
      audioBook = audioBookResponse.value.data;
    }

    if (transcriptResponse.status === "fulfilled") {
      transcript = transcriptResponse.value.data;
    }

    if (summaryResponse.status === "fulfilled") {
      summary = summaryResponse.value.data;
    }

    if (tagsResponse.status === "fulfilled") {
      tags = tagsResponse.value.data;
    }

    if (jobsResponse.status === "fulfilled") {
      processingJobs = jobsResponse.value.data || [];
    }
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
                    <div>
                      <h3 className="font-semibold text-lg mb-2">
                        Description
                      </h3>
                      <p className="text-muted-foreground">
                        {audioBook.description || "No description available."}
                      </p>
                    </div>

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

          {/* AI Processing Tabs */}
          <Card>
            <CardHeader>
              <CardTitle>AI Processing</CardTitle>
              <CardDescription>
                View and manage AI-generated content
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Tabs defaultValue="transcript" className="w-full">
                <TabsList className="grid w-full grid-cols-4">
                  <TabsTrigger value="transcript">Transcript</TabsTrigger>
                  <TabsTrigger value="summary">Summary</TabsTrigger>
                  <TabsTrigger value="tags">Tags</TabsTrigger>
                  <TabsTrigger value="jobs">Processing Jobs</TabsTrigger>
                </TabsList>

                <TabsContent value="transcript" className="space-y-4">
                  {transcript ? (
                    <div>
                      <div className="flex justify-between items-center mb-4">
                        <div>
                          <Badge variant="secondary">
                            Confidence:{" "}
                            {Math.round(transcript.confidence_score * 100)}%
                          </Badge>
                          <Badge variant="outline" className="ml-2">
                            {transcript.language}
                          </Badge>
                        </div>
                        <Button size="sm" variant="outline">
                          <FileText className="h-4 w-4 mr-2" />
                          Export
                        </Button>
                      </div>
                      <div className="bg-muted p-4 rounded-md max-h-64 overflow-y-auto">
                        <p className="text-sm whitespace-pre-wrap">
                          {transcript.content}
                        </p>
                      </div>
                    </div>
                  ) : (
                    <div className="text-center py-8">
                      <FileText className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
                      <p className="text-muted-foreground">
                        No transcript available
                      </p>
                      <Button className="mt-2" size="sm">
                        Generate Transcript
                      </Button>
                    </div>
                  )}
                </TabsContent>

                <TabsContent value="summary" className="space-y-4">
                  {summary ? (
                    <div>
                      <div className="flex justify-between items-center mb-4">
                        <Badge variant="secondary">
                          Model: {summary.model_used}
                        </Badge>
                      </div>
                      <div className="bg-muted p-4 rounded-md">
                        <p className="text-sm">{summary.content}</p>
                      </div>
                    </div>
                  ) : (
                    <div className="text-center py-8">
                      <Brain className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
                      <p className="text-muted-foreground">
                        No summary available
                      </p>
                      <Button className="mt-2" size="sm">
                        Generate Summary
                      </Button>
                    </div>
                  )}
                </TabsContent>

                <TabsContent value="tags" className="space-y-4">
                  {tags ? (
                    <div>
                      <div className="flex justify-between items-center mb-4">
                        <Badge variant="secondary">
                          Model: {tags.model_used}
                        </Badge>
                      </div>
                      <div className="flex flex-wrap gap-2">
                        {tags.content.map((tag: string, index: number) => (
                          <Badge key={index} variant="outline">
                            {tag}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  ) : (
                    <div className="text-center py-8">
                      <Tags className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
                      <p className="text-muted-foreground">No tags available</p>
                      <Button className="mt-2" size="sm">
                        Generate Tags
                      </Button>
                    </div>
                  )}
                </TabsContent>

                <TabsContent value="jobs" className="space-y-4">
                  {processingJobs.length > 0 ? (
                    <div className="space-y-2">
                      {processingJobs.map((job) => (
                        <div
                          key={job.id}
                          className="flex justify-between items-center p-3 border rounded-md"
                        >
                          <div>
                            <p className="font-medium capitalize">
                              {job.job_type}
                            </p>
                            <p className="text-sm text-muted-foreground">
                              {new Date(job.created_at).toLocaleString()}
                            </p>
                          </div>
                          <Badge
                            variant={
                              job.status === "completed"
                                ? "default"
                                : job.status === "failed"
                                ? "destructive"
                                : "secondary"
                            }
                          >
                            {job.status}
                          </Badge>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <div className="text-center py-8">
                      <p className="text-muted-foreground">
                        No processing jobs
                      </p>
                    </div>
                  )}
                </TabsContent>
              </Tabs>
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
                <Brain className="h-4 w-4 mr-2" />
                Process with AI
              </Button>
              <Button className="w-full" variant="outline" size="sm">
                <FileText className="h-4 w-4 mr-2" />
                Generate Transcript
              </Button>
              <Button className="w-full" variant="outline" size="sm">
                <Tags className="h-4 w-4 mr-2" />
                Generate Tags
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
