"use client";

import { useState, useEffect } from "react";
import { useParams, useRouter } from "next/navigation";
import { publicAPI, progressAPI, bookmarksAPI } from "@/lib/api";
import {
  Play,
  Pause,
  SkipBack,
  SkipForward,
  Volume2,
  Heart,
  Bookmark,
  Share2,
  Download,
  Clock,
  User,
  FileText,
  Brain,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { AudioPlayer } from "@/components/audio-player";

export default function AudioBookListeningPage() {
  const params = useParams();
  const router = useRouter();
  const [audioBook, setAudioBook] = useState<any>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [currentProgress, setCurrentProgress] = useState(0);
  const [isLiked, setIsLiked] = useState(false);
  const [isInLibrary, setIsInLibrary] = useState(false);
  const [transcript, setTranscript] = useState<any>(null);
  const [summary, setSummary] = useState<any>(null);
  const [bookmarks, setBookmarks] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchAudioBook = async () => {
      try {
        const response = await publicAPI.getPublicAudioBook(
          params.id as string
        );
        setAudioBook(response.data);

        // Fetch additional data
        const [transcriptRes, summaryRes, bookmarksRes] =
          await Promise.allSettled([
            fetch(`/api/transcripts/${params.id}`),
            fetch(`/api/summaries/${params.id}`),
            bookmarksAPI.getBookmarks(params.id as string),
          ]);

        if (transcriptRes.status === "fulfilled") {
          const transcriptData = await transcriptRes.value.json();
          setTranscript(transcriptData.data);
        }

        if (summaryRes.status === "fulfilled") {
          const summaryData = await summaryRes.value.json();
          setSummary(summaryData.data);
        }

        if (bookmarksRes.status === "fulfilled") {
          setBookmarks(bookmarksRes.value.data || []);
        }
      } catch (error) {
        console.error("Failed to fetch audio book:", error);
      } finally {
        setLoading(false);
      }
    };

    if (params.id) {
      fetchAudioBook();
    }
  }, [params.id]);

  const handlePlayPause = (playing: boolean) => {
    setIsPlaying(playing);
  };

  const handleProgressChange = (progress: number) => {
    setCurrentProgress(progress);
    // Save progress to backend
    progressAPI.updateProgress(params.id as string, { current_time: progress });
  };

  const toggleLike = () => {
    setIsLiked(!isLiked);
    // TODO: Implement like functionality
  };

  const toggleLibrary = () => {
    setIsInLibrary(!isInLibrary);
    // TODO: Implement library toggle
  };

  const addBookmark = () => {
    const newBookmark = {
      audiobook_id: params.id,
      position: currentProgress,
      note: `Bookmark at ${formatTime(currentProgress)}`,
    };

    bookmarksAPI.createBookmark(params.id as string, newBookmark);
    setBookmarks([...bookmarks, newBookmark]);
  };

  const formatTime = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = Math.floor(seconds % 60);

    if (hours > 0) {
      return `${hours}:${minutes.toString().padStart(2, "0")}:${secs
        .toString()
        .padStart(2, "0")}`;
    }
    return `${minutes}:${secs.toString().padStart(2, "0")}`;
  };

  const formatDuration = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours}h ${minutes}m`;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-muted-foreground">Loading audio book...</p>
        </div>
      </div>
    );
  }

  if (!audioBook) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <h2 className="text-2xl font-bold mb-2">Audio Book Not Found</h2>
          <p className="text-muted-foreground mb-4">
            The audio book you're looking for doesn't exist or has been removed.
          </p>
          <Button onClick={() => router.push("/library")}>
            Back to Library
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-start">
        <div>
          <h1 className="font-bold text-3xl mb-2">{audioBook.title}</h1>
          <p className="text-muted-foreground text-lg flex items-center gap-2">
            <User className="h-4 w-4" />
            {audioBook.author}
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={toggleLike}
            className={isLiked ? "text-red-600" : ""}
          >
            <Heart
              className={`h-4 w-4 mr-2 ${isLiked ? "fill-current" : ""}`}
            />
            {isLiked ? "Liked" : "Like"}
          </Button>
          <Button variant="outline" size="sm" onClick={toggleLibrary}>
            {isInLibrary ? "Remove from Library" : "Add to Library"}
          </Button>
          <Button variant="outline" size="sm">
            <Share2 className="h-4 w-4 mr-2" />
            Share
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Cover and Info */}
          <Card>
            <CardContent className="p-6">
              <div className="flex gap-6">
                {audioBook.cover_image && (
                  <img
                    src={audioBook.cover_image}
                    alt={audioBook.title}
                    className="w-48 h-72 object-cover rounded-md"
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
                          <p className="font-medium flex items-center gap-1">
                            <Clock className="h-4 w-4" />
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
                      <Button size="sm" variant="outline" onClick={addBookmark}>
                        <Bookmark className="h-4 w-4 mr-2" />
                        Add Bookmark
                      </Button>
                      <Button size="sm" variant="outline">
                        <Download className="h-4 w-4 mr-2" />
                        Download
                      </Button>
                    </div>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Audio Player */}
          <Card>
            <CardHeader>
              <CardTitle>Now Playing</CardTitle>
            </CardHeader>
            <CardContent>
              <AudioPlayer
                audioUrl={audioBook.file_url || ""}
                title={audioBook.title}
                author={audioBook.author}
                coverImage={audioBook.cover_image}
                duration={audioBook.duration}
                onProgressChange={handleProgressChange}
                onPlayPause={handlePlayPause}
              />
            </CardContent>
          </Card>

          {/* Content Tabs */}
          <Card>
            <CardHeader>
              <CardTitle>Content</CardTitle>
            </CardHeader>
            <CardContent>
              <Tabs defaultValue="transcript" className="w-full">
                <TabsList className="grid w-full grid-cols-3">
                  <TabsTrigger value="transcript">Transcript</TabsTrigger>
                  <TabsTrigger value="summary">Summary</TabsTrigger>
                  <TabsTrigger value="bookmarks">Bookmarks</TabsTrigger>
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
                      <div className="bg-muted p-4 rounded-md max-h-96 overflow-y-auto">
                        <p className="text-sm whitespace-pre-wrap leading-relaxed">
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
                    </div>
                  )}
                </TabsContent>

                <TabsContent value="summary" className="space-y-4">
                  {summary ? (
                    <div>
                      <div className="flex justify-between items-center mb-4">
                        <Badge variant="secondary">AI Generated Summary</Badge>
                      </div>
                      <div className="bg-muted p-4 rounded-md">
                        <p className="text-sm leading-relaxed">
                          {summary.content}
                        </p>
                      </div>
                    </div>
                  ) : (
                    <div className="text-center py-8">
                      <Brain className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
                      <p className="text-muted-foreground">
                        No summary available
                      </p>
                    </div>
                  )}
                </TabsContent>

                <TabsContent value="bookmarks" className="space-y-4">
                  {bookmarks.length > 0 ? (
                    <div className="space-y-2">
                      {bookmarks.map((bookmark, index) => (
                        <div
                          key={index}
                          className="flex justify-between items-center p-3 border rounded-md"
                        >
                          <div>
                            <p className="font-medium">{bookmark.note}</p>
                            <p className="text-sm text-muted-foreground">
                              {formatTime(bookmark.position)}
                            </p>
                          </div>
                          <Button size="sm" variant="outline">
                            Go to
                          </Button>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <div className="text-center py-8">
                      <Bookmark className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
                      <p className="text-muted-foreground">No bookmarks yet</p>
                      <p className="text-sm text-muted-foreground">
                        Add bookmarks while listening to mark important moments
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
          {/* Progress */}
          <Card>
            <CardHeader>
              <CardTitle>Your Progress</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                <div className="flex justify-between text-sm">
                  <span>Current Position</span>
                  <span>{formatTime(currentProgress)}</span>
                </div>
                <div className="w-full bg-muted rounded-full h-2">
                  <div
                    className="bg-primary h-2 rounded-full transition-all"
                    style={{
                      width: `${
                        audioBook.duration
                          ? (currentProgress / audioBook.duration) * 100
                          : 0
                      }%`,
                    }}
                  />
                </div>
                <div className="flex justify-between text-sm text-muted-foreground">
                  <span>0:00</span>
                  <span>{formatTime(audioBook.duration || 0)}</span>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Quick Actions */}
          <Card>
            <CardHeader>
              <CardTitle>Quick Actions</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <Button className="w-full" size="sm">
                <SkipBack className="h-4 w-4 mr-2" />
                Skip Back 30s
              </Button>
              <Button className="w-full" size="sm">
                <SkipForward className="h-4 w-4 mr-2" />
                Skip Forward 30s
              </Button>
              <Button className="w-full" variant="outline" size="sm">
                <Bookmark className="h-4 w-4 mr-2" />
                Add Bookmark
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
