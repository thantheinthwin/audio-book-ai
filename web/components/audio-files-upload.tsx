"use client";

import {
  AlertCircleIcon,
  DownloadIcon,
  FileArchiveIcon,
  FileIcon,
  FileSpreadsheetIcon,
  FileTextIcon,
  HeadphonesIcon,
  ImageIcon,
  Trash2Icon,
  UploadCloudIcon,
  UploadIcon,
  VideoIcon,
  Plus,
} from "lucide-react";

import { formatBytes, useFileUpload } from "@/hooks/use-file-upload";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useFieldArray, Control, UseFormRegister } from "react-hook-form";

interface ChapterField {
  id: string;
  chapter_number: number;
  title: string;
  audio_file?: File;
}

interface AudioFilesUploadProps {
  maxSizeMB?: number;
  maxFiles?: number;
  control: Control<any>;
  register: UseFormRegister<any>;
  name: string; // Field array name
  className?: string;
}

const getFileIcon = (file: { file: File | { type: string; name: string } }) => {
  const fileType = file.file instanceof File ? file.file.type : file.file.type;
  const fileName = file.file instanceof File ? file.file.name : file.file.name;

  if (
    fileType.includes("pdf") ||
    fileName.endsWith(".pdf") ||
    fileType.includes("word") ||
    fileName.endsWith(".doc") ||
    fileName.endsWith(".docx")
  ) {
    return <FileTextIcon className="size-4" />;
  } else if (
    fileType.includes("zip") ||
    fileType.includes("archive") ||
    fileName.endsWith(".zip") ||
    fileName.endsWith(".rar")
  ) {
    return <FileArchiveIcon className="size-4" />;
  } else if (
    fileType.includes("excel") ||
    fileName.endsWith(".xls") ||
    fileName.endsWith(".xlsx")
  ) {
    return <FileSpreadsheetIcon className="size-4" />;
  } else if (fileType.includes("video/")) {
    return <VideoIcon className="size-4" />;
  } else if (fileType.includes("audio/")) {
    return <HeadphonesIcon className="size-4" />;
  } else if (fileType.startsWith("image/")) {
    return <ImageIcon className="size-4" />;
  }
  return <FileIcon className="size-4" />;
};

export default function AudioFilesUpload({
  maxSizeMB = 100,
  maxFiles = 20,
  control,
  register,
  name,
  className = "",
}: AudioFilesUploadProps) {
  const maxSize = maxSizeMB * 1024 * 1024; // Convert MB to bytes

  const { fields, append, remove } = useFieldArray({
    control,
    name,
  });

  const [
    { files, isDragging, errors },
    {
      handleDragEnter,
      handleDragLeave,
      handleDragOver,
      handleDrop,
      openFileDialog,
      removeFile: removeUploadedFile,
      clearFiles,
      getInputProps,
    },
  ] = useFileUpload({
    multiple: true,
    maxFiles,
    maxSize,
    accept: "audio/*",
    onFilesChange: (uploadedFiles) => {
      // Clear existing fields and add new ones
      while (fields.length > 0) {
        remove(0);
      }

      // Add new fields for each uploaded file
      uploadedFiles.forEach((file, index) => {
        append({
          id: file.id,
          chapter_number: index + 1,
          title: "",
          audio_file:
            file.file instanceof File
              ? file.file
              : new File([], file.file.name),
        });
      });
    },
  });

  const handleAudioFileUpload = (
    event: React.ChangeEvent<HTMLInputElement>,
    index: number
  ) => {
    const file = event.target.files?.[0];
    if (file) {
      // Update the field with the new file
      const updatedFields = [...fields];
      updatedFields[index] = {
        ...updatedFields[index],
        audio_file: file,
      };

      // Replace the field
      remove(index);
      append(updatedFields[index], { shouldFocus: false });
    }
  };

  const removeChapter = (index: number) => {
    remove(index);
    // Reorder chapter numbers
    const updatedFields = fields.filter((_, i) => i !== index);
    updatedFields.forEach((field, i) => {
      const updatedField = {
        ...field,
        chapter_number: i + 1,
      };
      remove(i);
      append(updatedField, { shouldFocus: false });
    });
  };

  const addChapter = () => {
    const nextChapterNumber = fields.length + 1;
    append({
      id: `chapter-${Date.now()}-${Math.random()}`,
      chapter_number: nextChapterNumber,
      title: "",
      audio_file: undefined,
    });
  };

  return (
    <div className={`space-y-4 ${className}`}>
      {/* Drop area */}
      <div
        onDragEnter={handleDragEnter}
        onDragLeave={handleDragLeave}
        onDragOver={handleDragOver}
        onDrop={handleDrop}
        data-dragging={isDragging || undefined}
        data-files={fields.length > 0 || undefined}
        className="border-input data-[dragging=true]:bg-accent/50 has-[input:focus]:border-ring has-[input:focus]:ring-ring/50 flex min-h-56 flex-col items-center rounded-xl border border-dashed p-4 transition-colors not-data-[files]:justify-center has-[input:focus]:ring-[3px] data-[files]:hidden"
      >
        <input
          {...getInputProps()}
          className="sr-only"
          aria-label="Upload audio files"
        />
        <div
          className="bg-background mb-2 flex size-11 shrink-0 items-center justify-center rounded-full border"
          aria-hidden="true"
        >
          <HeadphonesIcon className="size-6" />
        </div>
        <p className="text-sm font-medium">Upload audio files</p>
        <p className="text-xs text-muted-foreground">
          Max {maxFiles} files ∙ Up to {formatBytes(maxSize)}
        </p>
        <Button
          variant="outline"
          className="mt-4"
          onClick={openFileDialog}
          type="button"
        >
          <UploadIcon
            className="-ms-0.5 size-3.5 opacity-60"
            aria-hidden="true"
          />
          Select files
        </Button>
      </div>

      {fields.length > 0 && (
        <>
          {/* Table with files */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold">
                Audio Files ({fields.length})
              </h3>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={addChapter}
                  type="button"
                >
                  <Plus
                    className="-ms-0.5 size-3.5 opacity-60"
                    aria-hidden="true"
                  />
                  Add Chapter
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={clearFiles}
                  type="button"
                >
                  <Trash2Icon
                    className="-ms-0.5 size-3.5 opacity-60"
                    aria-hidden="true"
                  />
                  Remove all
                </Button>
              </div>
            </div>

            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Chapter</TableHead>
                    <TableHead>Chapter Title</TableHead>
                    <TableHead>Audio File</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead>Size</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {fields.map((field, index) => (
                    <TableRow key={field.id}>
                      <TableCell className="font-medium">
                        Chapter {field.chapter_number}
                      </TableCell>
                      <TableCell>
                        <Input
                          placeholder={`Chapter ${field.chapter_number} title`}
                          {...register(`${name}.${index}.title`)}
                          className="w-full"
                        />
                      </TableCell>
                      <TableCell>
                        <Input
                          type="file"
                          accept="audio/*"
                          onChange={(e) => handleAudioFileUpload(e, index)}
                          className="w-full"
                        />
                        {field.audio_file && (
                          <div className="flex items-center gap-2 mt-1">
                            {getFileIcon({ file: field.audio_file })}
                            <span className="text-sm text-green-600">
                              {field.audio_file.name}
                            </span>
                          </div>
                        )}
                      </TableCell>
                      <TableCell>
                        {field.audio_file
                          ? field.audio_file.type
                              .split("/")[1]
                              ?.toUpperCase() || "UNKNOWN"
                          : "No file"}
                      </TableCell>
                      <TableCell>
                        {field.audio_file
                          ? formatBytes(field.audio_file.size)
                          : "-"}
                      </TableCell>
                      <TableCell>
                        <div className="flex gap-1">
                          {field.audio_file && (
                            <Button
                              size="icon"
                              variant="ghost"
                              className="text-muted-foreground/80 hover:text-foreground size-8 hover:bg-transparent"
                              aria-label={`Download ${field.audio_file.name}`}
                              onClick={() => {
                                const url = URL.createObjectURL(
                                  field.audio_file!
                                );
                                window.open(url, "_blank");
                              }}
                              type="button"
                            >
                              <DownloadIcon className="size-4" />
                            </Button>
                          )}
                          <Button
                            size="icon"
                            variant="ghost"
                            className="text-muted-foreground/80 hover:text-foreground size-8 hover:bg-transparent"
                            aria-label={`Remove chapter ${field.chapter_number}`}
                            onClick={() => removeChapter(index)}
                            type="button"
                          >
                            <Trash2Icon className="size-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </div>
        </>
      )}

      {errors.length > 0 && (
        <div
          className="text-destructive flex items-center gap-1 text-xs"
          role="alert"
        >
          <AlertCircleIcon className="size-3" />
          {errors[0]}
        </div>
      )}

      <p
        aria-live="polite"
        role="region"
        className="text-muted-foreground mt-2 text-center text-xs"
      >
        Multiple audio files uploader with React Hook Form field array
      </p>
    </div>
  );
}
