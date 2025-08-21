"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { createClient } from "@/lib/supabase/client";
import { useAudioBook, useAudioBookJobStatus } from "@/hooks/use-audiobooks";
import { notFound } from "next/navigation";
import {
  Play,
  Edit,
  Trash2,
  Download,
  Share2,
  Loader2,
  CheckCircle,
  AlertCircle,
  FileAudio,
  Brain,
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
import { Separator } from "@/components/ui/separator";
import { Progress } from "@/components/ui/progress";
import { Badge } from "@/components/ui/badge";

interface PageProps {
  params: {
    id: string;
  };
}

export default function AudioBookDetailPage({ params }: PageProps) {
  const router = useRouter();
  const {
    data: audioBookResponse,
    error: audioBookError,
    isLoading: audioBookLoading,
  } = useAudioBook(params.id);
  const { data: jobStatusResponse, error: jobStatusError } =
    useAudioBookJobStatus(params.id);

  const audioBook = audioBookResponse?.data;
  const jobStatus = jobStatusResponse?.data;

  // Check authentication
  useEffect(() => {
    const checkAuth = async () => {
      const supabase = createClient();
      const { data, error } = await supabase.auth.getSession();

      if (error || !data.session) {
        router.push("/auth/login");
        return;
      }

      // Check if user is admin
      const userRole = data.session.user.user_metadata.role || "user";
      if (userRole !== "admin") {
        router.push("/");
      }
    };

    checkAuth();
  }, [router]);

  if (audioBookLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  if (audioBookError || !audioBook) {
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

  const getJobStatusIcon = (jobType: string, status: string) => {
    switch (jobType) {
      case "transcribe":
        return <FileAudio className="h-4 w-4" />;
      case "summarize":
      case "tag":
      case "embed":
        return <Brain className="h-4 w-4" />;
      default:
        return <Loader2 className="h-4 w-4" />;
    }
  };

  const getJobStatusColor = (status: string) => {
    switch (status) {
      case "completed":
        return "bg-green-100 text-green-800";
      case "failed":
        return "bg-red-100 text-red-800";
      case "running":
        return "bg-blue-100 text-blue-800";
      case "pending":
        return "bg-yellow-100 text-yellow-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  const getJobTypeDisplayName = (jobType: string) => {
    switch (jobType) {
      case "transcribe":
        return "Transcribing";
      case "summarize":
        return "Summarizing";
      case "tag":
        return "Generating Tags";
      case "embed":
        return "Creating Embeddings";
      default:
        return jobType;
    }
  };

  const isProcessing = jobStatus?.overall_status === "processing";
  const isCompleted = jobStatus?.overall_status === "completed";
  const isFailed = jobStatus?.overall_status === "failed";

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

      {/* Processing Status */}
      {jobStatus && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              {isProcessing && <Loader2 className="h-5 w-5 animate-spin" />}
              {isCompleted && (
                <CheckCircle className="h-5 w-5 text-green-500" />
              )}
              {isFailed && <AlertCircle className="h-5 w-5 text-red-500" />}
              Processing Status
            </CardTitle>
            <CardDescription>
              {isProcessing &&
                "Your audiobook is being processed. This may take a few minutes."}
              {isCompleted && "Your audiobook has been successfully processed!"}
              {isFailed && "There was an error processing your audiobook."}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Overall Progress */}
            <div className="space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span>Overall Progress</span>
                <span>{Math.round((jobStatus.progress || 0) * 100)}%</span>
              </div>
              <Progress
                value={(jobStatus.progress || 0) * 100}
                className="w-full"
              />
            </div>

            {/* Job Details */}
            <div className="space-y-3">
              <h4 className="font-medium">Processing Jobs</h4>
              {jobStatus.jobs?.map((job) => (
                <div
                  key={job.id}
                  className="flex items-center justify-between p-3 border rounded-lg"
                >
                  <div className="flex items-center gap-3">
                    {getJobStatusIcon(job.job_type, job.status)}
                    <div>
                      <p className="font-medium">
                        {getJobTypeDisplayName(job.job_type)}
                      </p>
                      <p className="text-sm text-muted-foreground">
                        {job.started_at &&
                          `Started: ${new Date(
                            job.started_at
                          ).toLocaleTimeString()}`}
                        {job.completed_at &&
                          `Completed: ${new Date(
                            job.completed_at
                          ).toLocaleTimeString()}`}
                      </p>
                    </div>
                  </div>
                  <Badge className={getJobStatusColor(job.status)}>
                    {job.status}
                  </Badge>
                </div>
              ))}
            </div>

            {/* Job Statistics */}
            <div className="grid grid-cols-3 gap-4 text-sm">
              <div className="text-center">
                <div className="font-semibold text-green-600">
                  {jobStatus.completed_jobs || 0}
                </div>
                <div className="text-muted-foreground">Completed</div>
              </div>
              <div className="text-center">
                <div className="font-semibold text-red-600">
                  {jobStatus.failed_jobs || 0}
                </div>
                <div className="text-muted-foreground">Failed</div>
              </div>
              <div className="text-center">
                <div className="font-semibold text-blue-600">
                  {jobStatus.total_jobs || 0}
                </div>
                <div className="text-muted-foreground">Total</div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Cover Image and Basic Info */}
          <Card>
            <CardContent className="p-6">
              <div className="flex gap-6">
                {audioBook.cover_image_url && (
                  <img
                    src={audioBook.cover_image_url}
                    alt={audioBook.title}
                    className="w-32 h-48 object-cover rounded-md"
                  />
                )}
                <div className="flex-1">
                  <div className="space-y-4">
                    <div className="flex gap-4">
                      {audioBook.duration_seconds && (
                        <div>
                          <span className="text-sm text-muted-foreground">
                            Duration:
                          </span>
                          <p className="font-medium">
                            {formatDuration(audioBook.duration_seconds)}
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
                      <Button size="sm" disabled={isProcessing}>
                        <Play className="h-4 w-4 mr-2" />
                        Play
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        disabled={isProcessing}
                      >
                        <Download className="h-4 w-4 mr-2" />
                        Download
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        disabled={isProcessing}
                      >
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
              <Button className="w-full" size="sm" disabled={isProcessing}>
                <Play className="h-4 w-4 mr-2" />
                Play Audio Book
              </Button>
              <Button
                className="w-full"
                variant="outline"
                size="sm"
                disabled={isProcessing}
              >
                <Download className="h-4 w-4 mr-2" />
                Download
              </Button>
              <Button
                className="w-full"
                variant="outline"
                size="sm"
                disabled={isProcessing}
              >
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
                  {audioBook.cover_image_url || "Not available"}
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
