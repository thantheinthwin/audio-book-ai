# Supabase Storage Setup Guide

This guide will help you set up Supabase storage for the audio book AI application.

## Prerequisites

1. A Supabase project with storage enabled
2. Your Supabase project reference (found in your project URL)
3. Your Supabase service role key (found in Project Settings > API)

## Quick Setup

### Option 1: Use the Setup Script (Recommended)

```bash
# Run the setup script
./infra/setup-storage-env.sh
```

The script will prompt you for:

- Your Supabase project reference
- Your Supabase service role key
- Your desired bucket name (default: `audio`)

### Option 2: Manual Setup

1. **Set Environment Variables**

Add these to your `.env` file:

```bash
# Supabase Configuration
SUPABASE_URL=https://your-project-ref.supabase.co
SUPABASE_SECRET_KEY=your-service-role-key
SUPABASE_STORAGE_BUCKET=audio

# Supabase S3 Storage Configuration (for AWS S3 SDK)
SUPABASE_S3_ENDPOINT=https://your-project-ref.supabase.co/storage/v1/s3
SUPABASE_S3_REGION=us-east-1
SUPABASE_S3_ACCESS_KEY_ID=supabase
SUPABASE_S3_SECRET_KEY=your-service-role-key
```

2. **Create the Storage Bucket**

- Go to your Supabase dashboard: https://supabase.com/dashboard/project/your-project-ref/storage/buckets
- Click "Create a new bucket"
- Name it `audio` (or whatever you specified in `SUPABASE_STORAGE_BUCKET`)
- Set it to public if you want public access to uploaded files
- Click "Create bucket"

## Testing the Setup

### Test Storage Configuration

```bash
cd api
go run test_storage.go
```

This will:

- ✅ Check if the bucket exists
- 📤 Upload a test file
- 🌐 Generate public and signed URLs
- 🗑️ Delete the test file

### Test File Upload

```bash
cd api
go run example_storage.go
```

This will upload a sample image file to test the complete upload flow.

## Troubleshooting

### Common Issues

1. **"Bucket 'audio' does not exist or is not accessible"**

   - Make sure you've created the bucket in your Supabase dashboard
   - Verify the bucket name matches `SUPABASE_STORAGE_BUCKET`
   - Check that your service role key has the correct permissions

2. **"403 Forbidden" errors**

   - Ensure you're using the service role key, not the anon key
   - Verify your Supabase URL is correct
   - Check that storage is enabled in your Supabase project

3. **"Invalid endpoint" errors**

   - Make sure `SUPABASE_S3_ENDPOINT` follows the format: `https://your-project-ref.supabase.co/storage/v1/s3`
   - Verify your project reference is correct

4. **"Access denied" errors**
   - Check that your service role key is valid
   - Ensure the bucket is set to public if you need public access
   - Verify RLS (Row Level Security) policies if using private buckets

### Environment Variable Checklist

Make sure these are set correctly:

- ✅ `SUPABASE_URL` - Your Supabase project URL
- ✅ `SUPABASE_SECRET_KEY` - Your service role key (starts with 'eyJ')
- ✅ `SUPABASE_STORAGE_BUCKET` - Your bucket name (default: 'audio')
- ✅ `SUPABASE_S3_ENDPOINT` - S3 endpoint URL
- ✅ `SUPABASE_S3_REGION` - Region (usually 'us-east-1')
- ✅ `SUPABASE_S3_ACCESS_KEY_ID` - Should be 'supabase'
- ✅ `SUPABASE_S3_SECRET_KEY` - Same as your service role key

### Getting Your Supabase Credentials

1. **Project Reference**: Found in your project URL: `https://supabase.com/dashboard/project/your-project-ref`

2. **Service Role Key**:

   - Go to Project Settings > API
   - Copy the "service_role" key (not the "anon" key)

3. **Project URL**: `https://your-project-ref.supabase.co`

## Storage Bucket Configuration

### Public vs Private Buckets

- **Public Buckets**: Files are accessible via public URLs
- **Private Buckets**: Files require signed URLs or authentication

For this application, we recommend using a public bucket for audio files.

### RLS (Row Level Security) Policies

If you want to use a private bucket, you'll need to set up RLS policies. Example:

```sql
-- Allow authenticated users to upload files
CREATE POLICY "Users can upload files" ON storage.objects
FOR INSERT WITH CHECK (auth.role() = 'authenticated');

-- Allow users to read their own files
CREATE POLICY "Users can read own files" ON storage.objects
FOR SELECT USING (auth.uid()::text = (storage.foldername(name))[1]);
```

## File Types Supported

The application supports these audio file types:

- `.mp3` - MPEG Audio
- `.wav` - Waveform Audio
- `.m4a` - MPEG-4 Audio
- `.aac` - Advanced Audio Coding
- `.ogg` - Ogg Vorbis
- `.flac` - Free Lossless Audio Codec

## Next Steps

After setting up storage:

1. **Test the API**: Run the storage tests to verify everything works
2. **Upload Files**: Use the web interface to upload audio files
3. **Monitor Usage**: Check your Supabase dashboard for storage usage
4. **Set Up Backups**: Consider setting up automated backups for your storage bucket

## Support

If you're still having issues:

1. Check the Supabase documentation: https://supabase.com/docs/guides/storage
2. Verify your project settings in the Supabase dashboard
3. Test with the provided example scripts
4. Check the application logs for detailed error messages
