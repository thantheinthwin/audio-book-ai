/* eslint-disable @next/next/no-img-element */
"use client";

import { AlertCircleIcon, ImageIcon, UploadIcon, XIcon } from "lucide-react";
import { useFileUpload } from "@/hooks/use-file-upload";
import { Button } from "@/components/ui/button";

interface CoverImageUploadProps {
  maxSizeMB?: number;
  onImageChange?: (file: File | null) => void;
  className?: string;
}

export default function CoverImageUpload({
  maxSizeMB = 2,
  onImageChange,
  className = "",
}: CoverImageUploadProps) {
  const maxSize = maxSizeMB * 1024 * 1024; // Convert MB to bytes

  const [
    { files, isDragging, errors },
    {
      handleDragEnter,
      handleDragLeave,
      handleDragOver,
      handleDrop,
      openFileDialog,
      removeFile,
      getInputProps,
    },
  ] = useFileUpload({
    accept: "image/svg+xml,image/png,image/jpeg,image/jpg,image/gif",
    maxSize,
    onFilesChange: (files) => {
      // Call the callback with the first file (or null if no files)
      const file = files[0]?.file instanceof File ? files[0].file : null;
      onImageChange?.(file);
    },
  });

  const previewUrl = files[0]?.preview || null;
  const fileName = files[0]?.file.name || null;

  return (
    <div className={`relative ${className}`}>
      {/* Drop area */}
      <div
        onDragEnter={handleDragEnter}
        onDragLeave={handleDragLeave}
        onDragOver={handleDragOver}
        onDrop={handleDrop}
        data-dragging={isDragging || undefined}
        className="border-input data-[dragging=true]:bg-accent/50 has-[input:focus]:border-ring has-[input:focus]:ring-ring relative flex min-h-52 flex-col items-center justify-center overflow-hidden rounded-xl border border-dashed p-4 transition-colors has-[input:focus]:ring-[3px]"
      >
        <input
          {...getInputProps()}
          className="sr-only"
          aria-label="Upload cover image file"
        />
        {previewUrl ? (
          <div className="relative w-full h-full">
            <img
              src={previewUrl}
              alt={fileName || "uploaded image"}
              className="mx-auto max-h-full rounded object-contain"
            />
          </div>
        ) : (
          <div className="text-center">
            <div
              className="bg-background mb-2 flex size-11 shrink-0 items-center justify-center rounded-full border"
              aria-hidden="true"
            >
              <ImageIcon className="size-6" />
            </div>
            <p className="text-sm font-medium">Drop your cover image here</p>
            <p className="text-xs text-muted-foreground">
              SVG, PNG, JPG or GIF (max. {maxSizeMB}MB)
            </p>
            <Button
              variant="outline"
              className="mt-4"
              onClick={openFileDialog}
              type="button"
            >
              <UploadIcon
                className="-ms-1 size-4 opacity-60"
                aria-hidden="true"
              />
              Select image
            </Button>
          </div>
        )}
      </div>

      {/* Remove button */}
      {previewUrl && (
        <button
          type="button"
          className="focus-visible:border-ring focus-visible:ring-ring absolute right-2 top-2 z-50 flex size-8 cursor-pointer items-center justify-center rounded-full bg-black text-white transition-[color,box-shadow] outline-none hover:bg-black focus-visible:ring-[3px]"
          onClick={() => removeFile(files[0]?.id)}
          aria-label="Remove cover image"
        >
          <XIcon className="size-4" />
        </button>
      )}

      {/* Error messages */}
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
        Cover image uploader with drag & drop support
      </p>
    </div>
  );
}
