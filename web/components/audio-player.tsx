/* eslint-disable @next/next/no-img-element */
"use client";

import { useState, useRef, useEffect } from "react";
import {
  Play,
  Pause,
  SkipBack,
  SkipForward,
  Volume2,
  VolumeX,
  Maximize2,
  Minimize2,
  RotateCcw,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Slider } from "@/components/ui/slider";
import { Card, CardContent } from "@/components/ui/card";

interface AudioPlayerProps {
  audioUrl: string;
  title: string;
  author: string;
  coverImage?: string;
  duration?: number;
  onProgressChange?: (progress: number) => void;
  onPlayPause?: (isPlaying: boolean) => void;
  className?: string;
}

export function AudioPlayer({
  audioUrl,
  title,
  author,
  coverImage,
  duration,
  onProgressChange,
  onPlayPause,
  className = "",
}: AudioPlayerProps) {
  const audioRef = useRef<HTMLAudioElement>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration_, setDuration] = useState(0);
  const [volume, setVolume] = useState(1);
  const [isMuted, setIsMuted] = useState(false);
  const [isFullscreen, setIsFullscreen] = useState(false);

  useEffect(() => {
    const audio = audioRef.current;
    if (!audio) return;

    const updateTime = () => {
      setCurrentTime(audio.currentTime);
      onProgressChange?.(audio.currentTime);
    };

    const updateDuration = () => {
      setDuration(audio.duration);
    };

    const handleEnded = () => {
      setIsPlaying(false);
      onPlayPause?.(false);
    };

    audio.addEventListener("timeupdate", updateTime);
    audio.addEventListener("loadedmetadata", updateDuration);
    audio.addEventListener("ended", handleEnded);

    return () => {
      audio.removeEventListener("timeupdate", updateTime);
      audio.removeEventListener("loadedmetadata", updateDuration);
      audio.removeEventListener("ended", handleEnded);
    };
  }, [onProgressChange, onPlayPause]);

  useEffect(() => {
    if (duration) {
      setDuration(duration);
    }
  }, [duration]);

  const togglePlayPause = () => {
    const audio = audioRef.current;
    if (!audio) return;

    if (isPlaying) {
      audio.pause();
    } else {
      audio.play();
    }
    setIsPlaying(!isPlaying);
    onPlayPause?.(!isPlaying);
  };

  const skip = (seconds: number) => {
    const audio = audioRef.current;
    if (!audio) return;

    audio.currentTime = Math.max(
      0,
      Math.min(audio.currentTime + seconds, audio.duration)
    );
  };

  const seek = (value: number[]) => {
    const audio = audioRef.current;
    if (!audio) return;

    const newTime = (value[0] / 100) * audio.duration;
    audio.currentTime = newTime;
  };

  const handleVolumeChange = (value: number[]) => {
    const audio = audioRef.current;
    if (!audio) return;

    const newVolume = value[0] / 100;
    setVolume(newVolume);
    audio.volume = newVolume;
    setIsMuted(newVolume === 0);
  };

  const toggleMute = () => {
    const audio = audioRef.current;
    if (!audio) return;

    if (isMuted) {
      audio.volume = volume;
      setIsMuted(false);
    } else {
      audio.volume = 0;
      setIsMuted(true);
    }
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

  const progress = duration_ > 0 ? (currentTime / duration_) * 100 : 0;

  return (
    <Card
      className={`${className} ${
        isFullscreen ? "fixed inset-0 z-50 rounded-none" : ""
      }`}
    >
      <CardContent className="p-4">
        <div className="flex items-center gap-4">
          {/* Cover Image */}
          {coverImage && (
            <img
              src={coverImage}
              alt={title}
              className="w-16 h-16 object-cover rounded-md"
            />
          )}

          {/* Track Info */}
          <div className="flex-1 min-w-0">
            <h3 className="font-semibold text-sm truncate">{title}</h3>
            <p className="text-xs text-muted-foreground truncate">{author}</p>
          </div>

          {/* Controls */}
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => skip(-10)}
              className="h-8 w-8 p-0"
            >
              <SkipBack className="h-4 w-4" />
            </Button>

            <Button
              variant="ghost"
              size="sm"
              onClick={togglePlayPause}
              className="h-10 w-10 p-0"
            >
              {isPlaying ? (
                <Pause className="h-5 w-5" />
              ) : (
                <Play className="h-5 w-5" />
              )}
            </Button>

            <Button
              variant="ghost"
              size="sm"
              onClick={() => skip(10)}
              className="h-8 w-8 p-0"
            >
              <SkipForward className="h-4 w-4" />
            </Button>
          </div>

          {/* Progress */}
          <div className="flex-1 max-w-xs">
            <Slider
              value={[progress]}
              onValueChange={seek}
              max={100}
              step={0.1}
              className="w-full"
            />
            <div className="flex justify-between text-xs text-muted-foreground mt-1">
              <span>{formatTime(currentTime)}</span>
              <span>{formatTime(duration_)}</span>
            </div>
          </div>

          {/* Volume */}
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={toggleMute}
              className="h-8 w-8 p-0"
            >
              {isMuted ? (
                <VolumeX className="h-4 w-4" />
              ) : (
                <Volume2 className="h-4 w-4" />
              )}
            </Button>
            <Slider
              value={[isMuted ? 0 : volume * 100]}
              onValueChange={handleVolumeChange}
              max={100}
              className="w-20"
            />
          </div>

          {/* Fullscreen Toggle */}
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setIsFullscreen(!isFullscreen)}
            className="h-8 w-8 p-0"
          >
            {isFullscreen ? (
              <Minimize2 className="h-4 w-4" />
            ) : (
              <Maximize2 className="h-4 w-4" />
            )}
          </Button>
        </div>

        {/* Hidden Audio Element */}
        <audio
          ref={audioRef}
          src={audioUrl}
          preload="metadata"
          className="hidden"
        />
      </CardContent>
    </Card>
  );
}

// Mini Player Component
export function MiniAudioPlayer({
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  audioUrl,
  title,
  author,
  coverImage,
  isPlaying,
  onPlayPause,
  onClose,
}: {
  audioUrl: string;
  title: string;
  author: string;
  coverImage?: string;
  isPlaying: boolean;
  onPlayPause: () => void;
  onClose: () => void;
}) {
  return (
    <div className="fixed bottom-4 right-4 z-40">
      <Card className="w-80 shadow-lg">
        <CardContent className="p-3">
          <div className="flex items-center gap-3">
            {coverImage && (
              <img
                src={coverImage}
                alt={title}
                className="w-12 h-12 object-cover rounded-md"
              />
            )}

            <div className="flex-1 min-w-0">
              <h4 className="font-medium text-sm truncate">{title}</h4>
              <p className="text-xs text-muted-foreground truncate">{author}</p>
            </div>

            <div className="flex items-center gap-1">
              <Button
                variant="ghost"
                size="sm"
                onClick={onPlayPause}
                className="h-8 w-8 p-0"
              >
                {isPlaying ? (
                  <Pause className="h-4 w-4" />
                ) : (
                  <Play className="h-4 w-4" />
                )}
              </Button>

              <Button
                variant="ghost"
                size="sm"
                onClick={onClose}
                className="h-8 w-8 p-0"
              >
                <RotateCcw className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
