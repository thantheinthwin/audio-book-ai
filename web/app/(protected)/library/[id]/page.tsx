"use client";

import { useState, useRef } from "react";
import { useParams } from "next/navigation";
import { useAudioBook, useAudioBookJobStatus } from "@/hooks/use-audiobooks";
import { notFound } from "next/navigation";
import {
  Play,
  Pause,
  Edit,
  Trash2,
  Loader2,
  CheckCircle,
  AlertCircle,
  FileAudio,
  Brain,
  Bot,
  ShoppingCart,
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

  const {
    data: audioBookResponse,
    error: audioBookError,
    isLoading: audioBookLoading,
  } = useAudioBook(params.id as string);

  const audioBook = audioBookResponse?.data;

  console.log("audioBook", audioBook);

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
              {/* <Button variant="outline">
                <Edit className="h-4 w-4" />
                Edit
              </Button> */}
              <Button variant={"secondary"} className="w-full">
                <ShoppingCart className="h-4 w-4" />
                Add to cart
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
              <p className=" text-sm">{audioBook.summary}</p>
            </div>
            <div className="grid gap-1">
              <h2 className="text-muted-foreground text-sm">Tags</h2>
              <p className="text-xs">{audioBook.tags?.join(", ")}</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
