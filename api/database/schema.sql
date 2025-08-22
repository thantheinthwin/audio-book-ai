-- Audio Book AI Database Schema
-- This file contains all the SQL DDL statements to create the database schema

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable vector extension for embeddings (if using pgvector)
CREATE EXTENSION IF NOT EXISTS vector;

-- Audio Books table
CREATE TABLE IF NOT EXISTS audiobooks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    author VARCHAR(255) NOT NULL,
    summary TEXT,
    tags TEXT[],
    duration_seconds INTEGER,
    cover_image_url VARCHAR(500),
    language VARCHAR(2) NOT NULL,
    is_public BOOLEAN DEFAULT false,
    price DECIMAL(10,2) DEFAULT 0.00,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_by UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Uploads table
CREATE TABLE IF NOT EXISTS uploads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    upload_type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    total_files INTEGER NOT NULL,
    uploaded_files INTEGER DEFAULT 0,
    total_size_bytes BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Upload Files table
CREATE TABLE IF NOT EXISTS upload_files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    upload_id UUID NOT NULL REFERENCES uploads(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    chapter_number INTEGER,
    chapter_title VARCHAR(255),
    duration_seconds INTEGER,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    error TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Chapters table
CREATE TABLE IF NOT EXISTS chapters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    audiobook_id UUID NOT NULL REFERENCES audiobooks(id) ON DELETE CASCADE,
    upload_file_id UUID REFERENCES upload_files(id),
    chapter_number INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_url VARCHAR(500),
    file_size_bytes BIGINT,
    mime_type VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(audiobook_id, chapter_number)
);

-- Chapter Transcripts table
CREATE TABLE IF NOT EXISTS chapter_transcripts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chapter_id UUID NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    audiobook_id UUID NOT NULL REFERENCES audiobooks(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    segments JSONB,
    language VARCHAR(10),
    confidence_score DECIMAL(3,2),
    processing_time_seconds INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(chapter_id)
);

-- AI Outputs table
CREATE TABLE IF NOT EXISTS ai_outputs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    audiobook_id UUID NOT NULL REFERENCES audiobooks(id) ON DELETE CASCADE,
    output_type VARCHAR(20) NOT NULL,
    content JSONB NOT NULL,
    model_used VARCHAR(100),
    processing_time_seconds INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(audiobook_id, output_type)
);

-- Processing Jobs table
CREATE TABLE IF NOT EXISTS processing_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    audiobook_id UUID NOT NULL REFERENCES audiobooks(id) ON DELETE CASCADE,
    chapter_id UUID REFERENCES chapters(id) ON DELETE CASCADE,
    job_type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    redis_job_id VARCHAR(100),
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tags table
CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    category VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for better performance

-- Audio Books indexes
CREATE INDEX IF NOT EXISTS idx_audiobooks_created_by ON audiobooks(created_by);
CREATE INDEX IF NOT EXISTS idx_audiobooks_status ON audiobooks(status);
CREATE INDEX IF NOT EXISTS idx_audiobooks_language ON audiobooks(language);
CREATE INDEX IF NOT EXISTS idx_audiobooks_is_public ON audiobooks(is_public);
CREATE INDEX IF NOT EXISTS idx_audiobooks_created_at ON audiobooks(created_at);

-- Chapters indexes
CREATE INDEX IF NOT EXISTS idx_chapters_audiobook_id ON chapters(audiobook_id);
CREATE INDEX IF NOT EXISTS idx_chapters_upload_file_id ON chapters(upload_file_id);
CREATE INDEX IF NOT EXISTS idx_chapters_chapter_number ON chapters(chapter_number);

-- Chapter Transcripts indexes
CREATE INDEX IF NOT EXISTS idx_chapter_transcripts_chapter_id ON chapter_transcripts(chapter_id);
CREATE INDEX IF NOT EXISTS idx_chapter_transcripts_audiobook_id ON chapter_transcripts(audiobook_id);

-- AI Outputs indexes
CREATE INDEX IF NOT EXISTS idx_ai_outputs_audiobook_id ON ai_outputs(audiobook_id);
CREATE INDEX IF NOT EXISTS idx_ai_outputs_output_type ON ai_outputs(output_type);

-- Processing Jobs indexes
CREATE INDEX IF NOT EXISTS idx_processing_jobs_audiobook_id ON processing_jobs(audiobook_id);
CREATE INDEX IF NOT EXISTS idx_processing_jobs_status ON processing_jobs(status);
CREATE INDEX IF NOT EXISTS idx_processing_jobs_job_type ON processing_jobs(job_type);

-- Tags indexes
CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
CREATE INDEX IF NOT EXISTS idx_tags_category ON tags(category);

-- Uploads indexes
CREATE INDEX IF NOT EXISTS idx_uploads_user_id ON uploads(user_id);
CREATE INDEX IF NOT EXISTS idx_uploads_status ON uploads(status);
CREATE INDEX IF NOT EXISTS idx_uploads_upload_type ON uploads(upload_type);

-- Upload Files indexes
CREATE INDEX IF NOT EXISTS idx_upload_files_upload_id ON upload_files(upload_id);
CREATE INDEX IF NOT EXISTS idx_upload_files_status ON upload_files(status);
CREATE INDEX IF NOT EXISTS idx_upload_files_chapter_number ON upload_files(chapter_number);

-- Triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_audiobooks_updated_at BEFORE UPDATE ON audiobooks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_uploads_updated_at BEFORE UPDATE ON uploads
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_upload_files_updated_at BEFORE UPDATE ON upload_files
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Row Level Security (RLS) policies
-- Enable RLS on all tables
ALTER TABLE audiobooks ENABLE ROW LEVEL SECURITY;
ALTER TABLE chapters ENABLE ROW LEVEL SECURITY;
ALTER TABLE chapter_transcripts ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_outputs ENABLE ROW LEVEL SECURITY;
ALTER TABLE processing_jobs ENABLE ROW LEVEL SECURITY;
ALTER TABLE tags ENABLE ROW LEVEL SECURITY;
ALTER TABLE uploads ENABLE ROW LEVEL SECURITY;
ALTER TABLE upload_files ENABLE ROW LEVEL SECURITY;

-- Audio Books policies
CREATE POLICY "Users can view public audiobooks" ON audiobooks
    FOR SELECT USING (is_public = true);

CREATE POLICY "Users can view own audiobooks" ON audiobooks
    FOR SELECT USING (auth.uid() = created_by);

CREATE POLICY "Users can create own audiobooks" ON audiobooks
    FOR INSERT WITH CHECK (auth.uid() = created_by);

CREATE POLICY "Users can update own audiobooks" ON audiobooks
    FOR UPDATE USING (auth.uid() = created_by);

CREATE POLICY "Users can delete own audiobooks" ON audiobooks
    FOR DELETE USING (auth.uid() = created_by);

-- Admins can do everything (checking auth.users directly)
CREATE POLICY "Admins can do everything on audiobooks" ON audiobooks
    FOR ALL USING (
        EXISTS (
            SELECT 1 FROM auth.users 
            WHERE auth.users.id = auth.uid() 
            AND auth.users.raw_user_meta_data->>'role' = 'admin'
        )
    );

-- Uploads policies
CREATE POLICY "Users can view own uploads" ON uploads
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can create own uploads" ON uploads
    FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can update own uploads" ON uploads
    FOR UPDATE USING (auth.uid() = user_id);

CREATE POLICY "Users can delete own uploads" ON uploads
    FOR DELETE USING (auth.uid() = user_id);

-- Upload Files policies
CREATE POLICY "Users can view own upload files" ON upload_files
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM uploads 
            WHERE uploads.id = upload_files.upload_id 
            AND uploads.user_id = auth.uid()
        )
    );

CREATE POLICY "Users can create own upload files" ON upload_files
    FOR INSERT WITH CHECK (
        EXISTS (
            SELECT 1 FROM uploads 
            WHERE uploads.id = upload_files.upload_id 
            AND uploads.user_id = auth.uid()
        )
    );

CREATE POLICY "Users can update own upload files" ON upload_files
    FOR UPDATE USING (
        EXISTS (
            SELECT 1 FROM uploads 
            WHERE uploads.id = upload_files.upload_id 
            AND uploads.user_id = auth.uid()
        )
    );

CREATE POLICY "Users can delete own upload files" ON upload_files
    FOR DELETE USING (
        EXISTS (
            SELECT 1 FROM uploads 
            WHERE uploads.id = upload_files.upload_id 
            AND uploads.user_id = auth.uid()
        )
    );

-- Chapters policies
CREATE POLICY "Users can view chapters by audiobook ownership" ON chapters
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM audiobooks 
            WHERE audiobooks.id = chapters.audiobook_id 
            AND audiobooks.created_by = auth.uid()
        )
    );

CREATE POLICY "Users can view chapters by upload file ownership" ON chapters
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM upload_files 
            JOIN uploads ON uploads.id = upload_files.upload_id
            WHERE upload_files.id = chapters.upload_file_id 
            AND uploads.user_id = auth.uid()
        )
    );

CREATE POLICY "Users can create chapters by audiobook ownership" ON chapters
    FOR INSERT WITH CHECK (
        EXISTS (
            SELECT 1 FROM audiobooks 
            WHERE audiobooks.id = chapters.audiobook_id 
            AND audiobooks.created_by = auth.uid()
        )
    );

CREATE POLICY "Users can update chapters by audiobook ownership" ON chapters
    FOR UPDATE USING (
        EXISTS (
            SELECT 1 FROM audiobooks 
            WHERE audiobooks.id = chapters.audiobook_id 
            AND audiobooks.created_by = auth.uid()
        )
    );

CREATE POLICY "Users can delete chapters by audiobook ownership" ON chapters
    FOR DELETE USING (
        EXISTS (
            SELECT 1 FROM audiobooks 
            WHERE audiobooks.id = chapters.audiobook_id 
            AND audiobooks.created_by = auth.uid()
        )
    );

-- Similar policies for other tables...
-- (I'll create a separate file for all RLS policies to keep this clean)

-- Add comments to document the cascading delete behavior
COMMENT ON TABLE audiobooks IS 'Audiobooks table. When deleted, cascades to: chapters, transcripts, chapter_transcripts, ai_outputs, chapter_ai_outputs, processing_jobs, audiobook_tags, audiobook_embeddings, and related uploads via audiobook_uploads junction table.';

COMMENT ON TABLE uploads IS 'Uploads table. When deleted, cascades to: upload_files and chapters (via upload_file_id).';

COMMENT ON TABLE chapters IS 'Chapters table. When deleted, cascades to: chapter_transcripts and chapter_ai_outputs. References both audiobook_id and upload_file_id for proper tracking and cleanup.';

COMMENT ON COLUMN chapters.upload_file_id IS 'References the upload file that created this chapter. Allows tracking and cascading deletes from upload_files.';

-- Cart table for user shopping cart functionality
CREATE TABLE IF NOT EXISTS user_cart (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    audiobook_id UUID NOT NULL REFERENCES audiobooks(id) ON DELETE CASCADE,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, audiobook_id)
);

-- Cart indexes
CREATE INDEX IF NOT EXISTS idx_user_cart_user_id ON user_cart(user_id);
CREATE INDEX IF NOT EXISTS idx_user_cart_audiobook_id ON user_cart(audiobook_id);
CREATE INDEX IF NOT EXISTS idx_user_cart_added_at ON user_cart(added_at);

-- Cart RLS policies
ALTER TABLE user_cart ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view own cart items" ON user_cart
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can add items to own cart" ON user_cart
    FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can remove items from own cart" ON user_cart
    FOR DELETE USING (auth.uid() = user_id);

-- Purchased Audiobooks table for tracking user purchases
CREATE TABLE IF NOT EXISTS purchased_audiobooks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    audiobook_id UUID NOT NULL REFERENCES audiobooks(id) ON DELETE CASCADE,
    purchase_price DECIMAL(10,2) NOT NULL,
    purchased_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    transaction_id VARCHAR(255), -- For future payment integration
    payment_status VARCHAR(20) DEFAULT 'completed', -- completed, pending, failed, refunded
    UNIQUE(user_id, audiobook_id)
);

-- Purchased Audiobooks indexes
CREATE INDEX IF NOT EXISTS idx_purchased_audiobooks_user_id ON purchased_audiobooks(user_id);
CREATE INDEX IF NOT EXISTS idx_purchased_audiobooks_audiobook_id ON purchased_audiobooks(audiobook_id);
CREATE INDEX IF NOT EXISTS idx_purchased_audiobooks_purchased_at ON purchased_audiobooks(purchased_at);
CREATE INDEX IF NOT EXISTS idx_purchased_audiobooks_transaction_id ON purchased_audiobooks(transaction_id);

-- Purchased Audiobooks RLS policies
ALTER TABLE purchased_audiobooks ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view own purchased audiobooks" ON purchased_audiobooks
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can create own purchase records" ON purchased_audiobooks
    FOR INSERT WITH CHECK (auth.uid() = user_id);

-- Admins can view all purchases
CREATE POLICY "Admins can view all purchased audiobooks" ON purchased_audiobooks
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM auth.users 
            WHERE auth.users.id = auth.uid() 
            AND auth.users.raw_user_meta_data->>'role' = 'admin'
        )
    );

-- Insert some default tags
INSERT INTO tags (name, category) VALUES
    ('Fiction', 'Genre'),
    ('Non-Fiction', 'Genre'),
    ('Mystery', 'Genre'),
    ('Romance', 'Genre'),
    ('Science Fiction', 'Genre'),
    ('Fantasy', 'Genre'),
    ('Biography', 'Genre'),
    ('History', 'Genre'),
    ('Self-Help', 'Genre'),
    ('Business', 'Genre'),
    ('Technology', 'Genre'),
    ('Philosophy', 'Genre'),
    ('Psychology', 'Genre'),
    ('Education', 'Genre'),
    ('Entertainment', 'Genre')
ON CONFLICT (name) DO NOTHING;
