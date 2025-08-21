"use client";

import { useEffect, useState, useRef } from "react";
import { useParams, useRouter } from "next/navigation";
import { createClient } from "@/lib/supabase/client";
import { useAudioBook, useAudioBookJobStatus } from "@/hooks/use-audiobooks";
import { notFound } from "next/navigation";
import {
  Play,
  Pause,
  Edit,
  Trash2,
  Download,
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
import { Slider } from "@/components/ui/slider";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import Image from "next/image";

export default function AudioBookDetailPage() {
  const params = useParams();
  const router = useRouter();
  const [playingChapter, setPlayingChapter] = useState<string | null>(null);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [isJobStatusExpanded, setIsJobStatusExpanded] = useState(true);
  const audioRef = useRef<HTMLAudioElement | null>(null);

  const {
    data: audioBookResponse,
    error: audioBookError,
    isLoading: audioBookLoading,
  } = useAudioBook(params.id as string);
  const { data: jobStatusResponse } = useAudioBookJobStatus(
    params.id as string
  );

  const audioBook = audioBookResponse?.data;
  const jobStatus = jobStatusResponse?.data;

  console.log(jobStatus, audioBook);

  // Handle play/pause for a chapter
  const handlePlayPause = (chapterId: string, audioUrl: string) => {
    if (playingChapter === chapterId) {
      // Pause current chapter
      if (audioRef.current) {
        audioRef.current.pause();
        audioRef.current = null;
      }
      setPlayingChapter(null);
      setCurrentTime(0);
      setDuration(0);
    } else {
      // Always stop any currently playing audio first
      if (audioRef.current) {
        audioRef.current.pause();
        audioRef.current = null;
      }

      // Reset playing state and time
      setPlayingChapter(null);
      setCurrentTime(0);
      setDuration(0);

      // Play new chapter
      const audio = new Audio(audioUrl);
      audioRef.current = audio;

      // Add event listeners
      const handleEnded = () => {
        setPlayingChapter(null);
        audioRef.current = null;
        setCurrentTime(0);
        setDuration(0);
      };

      const handlePause = () => {
        setPlayingChapter(null);
        audioRef.current = null;
      };

      const handleError = () => {
        console.error("Error playing audio:", audio.error);
        setPlayingChapter(null);
        audioRef.current = null;
        setCurrentTime(0);
        setDuration(0);
      };

      const handleLoadedMetadata = () => {
        setDuration(audio.duration);
      };

      const handleTimeUpdate = () => {
        setCurrentTime(audio.currentTime);
      };

      audio.addEventListener("ended", handleEnded);
      audio.addEventListener("pause", handlePause);
      audio.addEventListener("error", handleError);
      audio.addEventListener("loadedmetadata", handleLoadedMetadata);
      audio.addEventListener("timeupdate", handleTimeUpdate);

      audio
        .play()
        .then(() => {
          setPlayingChapter(chapterId);
        })
        .catch((error) => {
          console.error("Error playing audio:", error);
          setPlayingChapter(null);
          audioRef.current = null;
          // Remove event listeners on error
          audio.removeEventListener("ended", handleEnded);
          audio.removeEventListener("pause", handlePause);
          audio.removeEventListener("error", handleError);
          audio.removeEventListener("loadedmetadata", handleLoadedMetadata);
          audio.removeEventListener("timeupdate", handleTimeUpdate);
        });
    }
  };

  // Handle slider change
  const handleSliderChange = (value: number[]) => {
    if (audioRef.current && duration > 0) {
      const newTime = (value[0] / 100) * duration;
      audioRef.current.currentTime = newTime;
      setCurrentTime(newTime);
    }
  };

  // Handle download
  const handleDownload = (audioUrl: string, chapterTitle: string) => {
    const link = document.createElement("a");
    link.href = audioUrl;
    link.download = `${chapterTitle}.mp3`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (audioRef.current) {
        audioRef.current.pause();
        audioRef.current = null;
      }
      setPlayingChapter(null);
      setCurrentTime(0);
      setDuration(0);
    };
  }, []);

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

  const getJobStatusIcon = (jobType: string) => {
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
            <div className="grid gap-1">
              <h2 className="text-muted-foreground text-sm">Tags</h2>
              <p className="text-xs">{audioBook.tags?.join(", ")}</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Processing Status */}
      {jobStatus && (
        <Card>
          <CardHeader
            className="flex flex-row justify-between items-start cursor-pointer hover:bg-muted/50 transition-colors"
            onClick={() => setIsJobStatusExpanded(!isJobStatusExpanded)}
          >
            <div className="space-y-2 flex-1">
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center gap-2">
                  {isProcessing && <Loader2 className="h-5 w-5 animate-spin" />}
                  {isCompleted && (
                    <CheckCircle className="h-5 w-5 text-green-500" />
                  )}
                  {isFailed && <AlertCircle className="h-5 w-5 text-red-500" />}
                  Processing Status
                </CardTitle>
                {/* <div className="flex items-center gap-2">
                  <span className="text-sm text-muted-foreground">
                    {isJobStatusExpanded
                      ? "Click to collapse"
                      : "Click to expand"}
                  </span>
                  {isJobStatusExpanded ? (
                    <ChevronUp className="h-4 w-4" />
                  ) : (
                    <ChevronDown className="h-4 w-4" />
                  )}
                </div> */}
              </div>
              <CardDescription>
                {isProcessing &&
                  "Your audiobook is being processed. This may take a few minutes."}
                {isCompleted &&
                  "Your audiobook has been successfully processed!"}
                {isFailed && "There was an error processing your audiobook."}
              </CardDescription>
            </div>
            {/* Overall Progress */}
            <div className="space-y-2 w-1/2">
              <div className="flex items-center justify-between text-sm">
                <span>Overall Progress</span>
                <span>{Math.round((jobStatus.progress || 0) * 100)}%</span>
              </div>
              <Progress
                value={(jobStatus.progress || 0) * 100}
                className="w-full h-2"
              />
            </div>
          </CardHeader>
          {isJobStatusExpanded && (
            <CardContent className="space-y-4 pt-4">
              {/* Job Details */}
              <div className="space-y-3">
                <h4 className="font-medium">Processing Jobs</h4>
                {jobStatus.jobs?.map((job) => (
                  <div
                    key={job.id}
                    className="flex items-center justify-between p-3 border rounded-lg"
                  >
                    <div className="flex items-center gap-3">
                      {getJobStatusIcon(job.job_type)}
                      <div>
                        <p className="font-medium">
                          {getJobTypeDisplayName(job.job_type)}
                        </p>
                        <div className="flex gap-1">
                          <p className="text-xs text-muted-foreground">
                            {job.started_at &&
                              `Started: ${new Date(
                                job.started_at
                              ).toLocaleTimeString()}`}{" "}
                          </p>
                          <p className="text-xs text-muted-foreground">
                            {job.completed_at &&
                              `Completed: ${new Date(
                                job.completed_at
                              ).toLocaleTimeString()}`}
                          </p>
                        </div>
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
          )}
        </Card>
      )}

      {/* Chapters Table */}
      {isCompleted && audioBook && (
        <Card>
          <CardHeader>
            <CardTitle>Chapters</CardTitle>
            <CardDescription>
              Listen to the chapters of your audiobook
            </CardDescription>
          </CardHeader>
          <CardContent>
            {audioBook.chapters && audioBook.chapters.length > 0 ? (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-16">No</TableHead>
                    <TableHead>Title</TableHead>
                    <TableHead className="w-64">Track</TableHead>
                    <TableHead className="w-32">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {audioBook.chapters.map((chapter) => (
                    <TableRow key={chapter.id}>
                      <TableCell className="font-medium">
                        {chapter.chapter_number}
                      </TableCell>
                      <TableCell>{chapter.title}</TableCell>
                      <TableCell>
                        <div className="space-y-2">
                          <Slider
                            value={[
                              playingChapter === chapter.id && duration > 0
                                ? (currentTime / duration) * 100
                                : 0,
                            ]}
                            onValueChange={(value) => {
                              if (
                                chapter.file_url &&
                                playingChapter === chapter.id
                              ) {
                                handleSliderChange(value);
                              }
                            }}
                            max={100}
                            step={0.1}
                            className="w-full"
                            disabled={!chapter.file_url}
                          />
                          <div className="flex justify-between text-xs text-muted-foreground">
                            <span>
                              {playingChapter === chapter.id
                                ? formatDuration(currentTime)
                                : "0:00:00"}
                            </span>
                            <span>
                              {chapter.duration_seconds
                                ? formatDuration(chapter.duration_seconds)
                                : "N/A"}
                            </span>
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="flex gap-2">
                          {chapter.file_url && (
                            <Button
                              variant="outline"
                              size="icon"
                              onClick={() =>
                                handlePlayPause(chapter.id, chapter.file_url!)
                              }
                            >
                              {playingChapter === chapter.id ? (
                                <Pause className="h-4 w-4" />
                              ) : (
                                <Play className="h-4 w-4" />
                              )}
                            </Button>
                          )}
                          {chapter.file_url && (
                            <Button
                              variant="outline"
                              size="icon"
                              onClick={() =>
                                handleDownload(chapter.file_url!, chapter.title)
                              }
                            >
                              <Download className="h-4 w-4" />
                            </Button>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            ) : (
              <div className="text-center py-8">
                <FileAudio className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
                <p className="text-muted-foreground">
                  No chapters available for this audiobook.
                </p>
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
