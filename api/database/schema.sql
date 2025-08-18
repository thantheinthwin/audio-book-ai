-- Audio Book AI Database Schema
-- This file contains all the SQL DDL statements to create the database schema

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable vector extension for embeddings (if using pgvector)
CREATE EXTENSION IF NOT EXISTS vector;

-- Users table (references Supabase auth.users)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL UNIQUE,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Audio Books table
CREATE TABLE IF NOT EXISTS audiobooks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    author VARCHAR(255) NOT NULL,
    description TEXT,
    duration_seconds INTEGER,
    file_size_bytes BIGINT,
    file_path VARCHAR(500) NOT NULL,
    file_url VARCHAR(500),
    cover_image_url VARCHAR(500),
    language VARCHAR(2) NOT NULL,
    is_public BOOLEAN DEFAULT false,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Chapters table
CREATE TABLE IF NOT EXISTS chapters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    audiobook_id UUID NOT NULL REFERENCES audiobooks(id) ON DELETE CASCADE,
    chapter_number INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    start_time_seconds INTEGER,
    end_time_seconds INTEGER,
    duration_seconds INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(audiobook_id, chapter_number)
);

-- Transcripts table
CREATE TABLE IF NOT EXISTS transcripts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    audiobook_id UUID NOT NULL REFERENCES audiobooks(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    segments JSONB,
    language VARCHAR(10),
    confidence_score DECIMAL(3,2),
    processing_time_seconds INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(audiobook_id)
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
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Chapter AI Outputs table
CREATE TABLE IF NOT EXISTS chapter_ai_outputs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chapter_id UUID NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    audiobook_id UUID NOT NULL REFERENCES audiobooks(id) ON DELETE CASCADE,
    output_type VARCHAR(20) NOT NULL,
    content JSONB NOT NULL,
    model_used VARCHAR(100),
    processing_time_seconds INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Processing Jobs table
CREATE TABLE IF NOT EXISTS processing_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    audiobook_id UUID NOT NULL REFERENCES audiobooks(id) ON DELETE CASCADE,
    job_type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    redis_job_id VARCHAR(100),
    error_message TEXT,
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

-- Audio Book Tags junction table
CREATE TABLE IF NOT EXISTS audiobook_tags (
    audiobook_id UUID NOT NULL REFERENCES audiobooks(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    confidence_score DECIMAL(3,2),
    is_ai_generated BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (audiobook_id, tag_id)
);

-- Audio Book Embeddings table
CREATE TABLE IF NOT EXISTS audiobook_embeddings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    audiobook_id UUID NOT NULL REFERENCES audiobooks(id) ON DELETE CASCADE,
    embedding vector(1536), -- OpenAI embedding dimension
    embedding_type VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Uploads table
CREATE TABLE IF NOT EXISTS uploads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
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
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    error TEXT,
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
CREATE INDEX IF NOT EXISTS idx_chapters_chapter_number ON chapters(chapter_number);

-- Transcripts indexes
CREATE INDEX IF NOT EXISTS idx_transcripts_audiobook_id ON transcripts(audiobook_id);

-- Chapter Transcripts indexes
CREATE INDEX IF NOT EXISTS idx_chapter_transcripts_chapter_id ON chapter_transcripts(chapter_id);
CREATE INDEX IF NOT EXISTS idx_chapter_transcripts_audiobook_id ON chapter_transcripts(audiobook_id);

-- AI Outputs indexes
CREATE INDEX IF NOT EXISTS idx_ai_outputs_audiobook_id ON ai_outputs(audiobook_id);
CREATE INDEX IF NOT EXISTS idx_ai_outputs_output_type ON ai_outputs(output_type);

-- Chapter AI Outputs indexes
CREATE INDEX IF NOT EXISTS idx_chapter_ai_outputs_chapter_id ON chapter_ai_outputs(chapter_id);
CREATE INDEX IF NOT EXISTS idx_chapter_ai_outputs_audiobook_id ON chapter_ai_outputs(audiobook_id);
CREATE INDEX IF NOT EXISTS idx_chapter_ai_outputs_output_type ON chapter_ai_outputs(output_type);

-- Processing Jobs indexes
CREATE INDEX IF NOT EXISTS idx_processing_jobs_audiobook_id ON processing_jobs(audiobook_id);
CREATE INDEX IF NOT EXISTS idx_processing_jobs_status ON processing_jobs(status);
CREATE INDEX IF NOT EXISTS idx_processing_jobs_job_type ON processing_jobs(job_type);

-- Tags indexes
CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
CREATE INDEX IF NOT EXISTS idx_tags_category ON tags(category);

-- Audio Book Tags indexes
CREATE INDEX IF NOT EXISTS idx_audiobook_tags_audiobook_id ON audiobook_tags(audiobook_id);
CREATE INDEX IF NOT EXISTS idx_audiobook_tags_tag_id ON audiobook_tags(tag_id);

-- Audio Book Embeddings indexes
CREATE INDEX IF NOT EXISTS idx_audiobook_embeddings_audiobook_id ON audiobook_embeddings(audiobook_id);
CREATE INDEX IF NOT EXISTS idx_audiobook_embeddings_type ON audiobook_embeddings(embedding_type);

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

-- Row Level Security (RLS) policies
-- Enable RLS on all tables
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE audiobooks ENABLE ROW LEVEL SECURITY;
ALTER TABLE chapters ENABLE ROW LEVEL SECURITY;
ALTER TABLE transcripts ENABLE ROW LEVEL SECURITY;
ALTER TABLE chapter_transcripts ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_outputs ENABLE ROW LEVEL SECURITY;
ALTER TABLE chapter_ai_outputs ENABLE ROW LEVEL SECURITY;
ALTER TABLE processing_jobs ENABLE ROW LEVEL SECURITY;
ALTER TABLE tags ENABLE ROW LEVEL SECURITY;
ALTER TABLE audiobook_tags ENABLE ROW LEVEL SECURITY;
ALTER TABLE audiobook_embeddings ENABLE ROW LEVEL SECURITY;
ALTER TABLE uploads ENABLE ROW LEVEL SECURITY;
ALTER TABLE upload_files ENABLE ROW LEVEL SECURITY;

-- Users can only see their own data
CREATE POLICY "Users can view own data" ON users
    FOR SELECT USING (auth.uid() = id);

CREATE POLICY "Users can update own data" ON users
    FOR UPDATE USING (auth.uid() = id);

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

-- Admins can do everything
CREATE POLICY "Admins can do everything on audiobooks" ON audiobooks
    FOR ALL USING (
        EXISTS (
            SELECT 1 FROM users 
            WHERE users.id = auth.uid() 
            AND users.role = 'admin'
        )
    );

-- Similar policies for other tables...
-- (I'll create a separate file for all RLS policies to keep this clean)

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
