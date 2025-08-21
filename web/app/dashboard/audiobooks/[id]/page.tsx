"use client";

import { useEffect } from "react";
import { useParams, useRouter } from "next/navigation";
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
  Bot,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Badge } from "@/components/ui/badge";
import Image from "next/image";

export default function AudioBookDetailPage() {
  const params = useParams();
  const router = useRouter();

  const {
    data: audioBookResponse,
    error: audioBookError,
    isLoading: audioBookLoading,
  } = useAudioBook(params.id as string);
  const { data: jobStatusResponse, error: jobStatusError } =
    useAudioBookJobStatus(params.id as string);

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
      <Card>
        <CardContent className="p-6 flex gap-4">
          <div className="flex flex-col gap-4">
            <Image
              src={"/images/placeholder.png"}
              alt={audioBook.title}
              width={100}
              height={100}
              className="w-48 h-48 object-cover rounded-md"
            />
            <div className="flex gap-2">
              <Button variant="outline">
                <Edit className="h-4 w-4" />
                Edit
              </Button>
              <Button variant={"destructive"}>
                <Trash2 className="h-4 w-4" />
                Delete
              </Button>
            </div>
          </div>
          <div className="flex flex-col flex-1 gap-2">
            <div className="grid gap-1">
              <h2 className="text-muted-foreground text-sm">Title</h2>
              <p className="font-semibold text-lg">{audioBook.title}</p>
            </div>
            <div className="grid gap-1">
              <h2 className="text-muted-foreground text-sm">Author</h2>
              <p className="font-semibold text-lg">{audioBook.author}</p>
            </div>
            <div className="grid gap-1">
              <h2 className="text-muted-foreground text-sm">Summary</h2>
              {jobStatus?.overall_status === "processing" ? (
                jobStatus.completed_jobs + jobStatus.failed_jobs ===
                jobStatus.total_jobs - 1 ? (
                  <div className="flex gap-1 items-center">
                    <Bot className="h-4 w-4" />
                    <p className="text-sm">Generating summary...</p>
                  </div>
                ) : (
                  <div className="flex gap-1 items-center">
                    Summary will be generated once all chapters are
                    transcribed...
                  </div>
                )
              ) : (
                <div className="flex flex-col gap-2">
                  <p className="text-sm">
                    {audioBook.summary || "No summary available"}
                  </p>
                </div>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

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
    </div>
  );
}
