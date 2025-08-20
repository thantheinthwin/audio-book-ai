#!/bin/bash

# Supabase Storage Environment Setup Script
# This script helps you set up the required environment variables for Supabase storage

echo "🔧 Supabase Storage Environment Setup"
echo "======================================"
echo ""

# Get project reference
read -p "Enter your Supabase project reference (e.g., abcdefghijklmnop): " PROJECT_REF

if [ -z "$PROJECT_REF" ]; then
    echo "❌ Project reference is required"
    exit 1
fi

# Get service role key
read -p "Enter your Supabase service role key (starts with 'eyJ'): " SERVICE_ROLE_KEY

if [ -z "$SERVICE_ROLE_KEY" ]; then
    echo "❌ Service role key is required"
    exit 1
fi

# Get bucket name
read -p "Enter your storage bucket name (default: audio): " BUCKET_NAME
BUCKET_NAME=${BUCKET_NAME:-audio}

echo ""
echo "📝 Generated environment variables:"
echo "===================================="
echo ""
echo "# Supabase Configuration"
echo "SUPABASE_URL=https://${PROJECT_REF}.supabase.co"
echo "SUPABASE_SECRET_KEY=${SERVICE_ROLE_KEY}"
echo "SUPABASE_STORAGE_BUCKET=${BUCKET_NAME}"
echo ""
echo "# Supabase S3 Storage Configuration (for AWS S3 SDK)"
echo "SUPABASE_S3_ENDPOINT=https://${PROJECT_REF}.supabase.co/storage/v1/s3"
echo "SUPABASE_S3_REGION=us-east-1"
echo "SUPABASE_S3_ACCESS_KEY_ID=supabase"
echo "SUPABASE_S3_SECRET_KEY=${SERVICE_ROLE_KEY}"
echo ""

# Ask if user wants to save to .env file
read -p "Do you want to save these to a .env file? (y/n): " SAVE_ENV

if [ "$SAVE_ENV" = "y" ] || [ "$SAVE_ENV" = "Y" ]; then
    ENV_FILE=".env"
    
    # Create or append to .env file
    {
        echo "# Supabase Configuration"
        echo "SUPABASE_URL=https://${PROJECT_REF}.supabase.co"
        echo "SUPABASE_SECRET_KEY=${SERVICE_ROLE_KEY}"
        echo "SUPABASE_STORAGE_BUCKET=${BUCKET_NAME}"
        echo ""
        echo "# Supabase S3 Storage Configuration (for AWS S3 SDK)"
        echo "SUPABASE_S3_ENDPOINT=https://${PROJECT_REF}.supabase.co/storage/v1/s3"
        echo "SUPABASE_S3_REGION=us-east-1"
        echo "SUPABASE_S3_ACCESS_KEY_ID=supabase"
        echo "SUPABASE_S3_SECRET_KEY=${SERVICE_ROLE_KEY}"
    } > "$ENV_FILE"
    
    echo "✅ Environment variables saved to $ENV_FILE"
    echo ""
    echo "📋 Next steps:"
    echo "1. Create the '${BUCKET_NAME}' bucket in your Supabase dashboard"
    echo "2. Set the bucket to public if you want public access"
    echo "3. Test the configuration by running: cd api && go run test_storage.go"
else
    echo ""
    echo "📋 Next steps:"
    echo "1. Copy the environment variables above to your .env file"
    echo "2. Create the '${BUCKET_NAME}' bucket in your Supabase dashboard"
    echo "3. Set the bucket to public if you want public access"
    echo "4. Test the configuration by running: cd api && go run test_storage.go"
fi

echo ""
echo "🔗 Supabase Dashboard: https://supabase.com/dashboard/project/${PROJECT_REF}/storage/buckets"
echo ""
