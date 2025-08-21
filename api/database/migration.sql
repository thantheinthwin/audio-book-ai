-- Migration to handle multiple upload sessions per audiobook
-- This allows adding new chapters to existing audiobooks

-- Create audiobook_uploads junction table to track multiple upload sessions per audiobook
CREATE TABLE IF NOT EXISTS audiobook_uploads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    audiobook_id UUID NOT NULL REFERENCES audiobooks(id) ON DELETE CASCADE,
    upload_id UUID NOT NULL REFERENCES uploads(id) ON DELETE CASCADE,
    upload_type VARCHAR(20) NOT NULL, -- 'initial', 'additional_chapters', etc.
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(audiobook_id, upload_id)
);

-- Add indexes for the new table
CREATE INDEX IF NOT EXISTS idx_audiobook_uploads_audiobook_id ON audiobook_uploads(audiobook_id);
CREATE INDEX IF NOT EXISTS idx_audiobook_uploads_upload_id ON audiobook_uploads(upload_id);
CREATE INDEX IF NOT EXISTS idx_audiobook_uploads_upload_type ON audiobook_uploads(upload_type);

-- Enable RLS on the new table
ALTER TABLE audiobook_uploads ENABLE ROW LEVEL SECURITY;

-- Add RLS policies for audiobook_uploads
CREATE POLICY "Users can view audiobook uploads by audiobook ownership" ON audiobook_uploads
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM audiobooks 
            WHERE audiobooks.id = audiobook_uploads.audiobook_id 
            AND audiobooks.created_by = auth.uid()
        )
    );

CREATE POLICY "Users can view audiobook uploads by upload ownership" ON audiobook_uploads
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM uploads 
            WHERE uploads.id = audiobook_uploads.upload_id 
            AND uploads.user_id = auth.uid()
        )
    );

CREATE POLICY "Users can create audiobook uploads by audiobook ownership" ON audiobook_uploads
    FOR INSERT WITH CHECK (
        EXISTS (
            SELECT 1 FROM audiobooks 
            WHERE audiobooks.id = audiobook_uploads.audiobook_id 
            AND audiobooks.created_by = auth.uid()
        )
    );

CREATE POLICY "Users can delete audiobook uploads by audiobook ownership" ON audiobook_uploads
    FOR DELETE USING (
        EXISTS (
            SELECT 1 FROM audiobooks 
            WHERE audiobooks.id = audiobook_uploads.audiobook_id 
            AND audiobooks.created_by = auth.uid()
        )
    );

-- Add comment to document the new relationship
COMMENT ON TABLE audiobook_uploads IS 'Junction table to track multiple upload sessions per audiobook. Allows adding new chapters to existing audiobooks while maintaining proper relationships.';

-- Migration to add cascading deletes from audiobooks to uploads and upload_files
-- This ensures that when an audiobook is deleted, related uploads and upload_files are also deleted

-- Add a trigger function to handle cascading deletes from audiobooks
CREATE OR REPLACE FUNCTION delete_audiobook_related_uploads()
RETURNS TRIGGER AS $$
BEGIN
    -- Delete uploads that are only associated with this audiobook
    DELETE FROM uploads 
    WHERE id IN (
        SELECT au.upload_id 
        FROM audiobook_uploads au 
        WHERE au.audiobook_id = OLD.id
        AND NOT EXISTS (
            SELECT 1 FROM audiobook_uploads au2 
            WHERE au2.upload_id = au.upload_id 
            AND au2.audiobook_id != OLD.id
        )
    );
    
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically delete related uploads when audiobook is deleted
DROP TRIGGER IF EXISTS trigger_delete_audiobook_related_uploads ON audiobooks;
CREATE TRIGGER trigger_delete_audiobook_related_uploads
    BEFORE DELETE ON audiobooks
    FOR EACH ROW
    EXECUTE FUNCTION delete_audiobook_related_uploads();

-- Add comment to document the cascading delete behavior
COMMENT ON FUNCTION delete_audiobook_related_uploads() IS 'Trigger function to delete uploads and upload_files when an audiobook is deleted. Only deletes uploads that are exclusively associated with the deleted audiobook.';

-- Add comment to document the simplified relationship
COMMENT ON TABLE audiobooks IS 'Audiobooks table. Related uploads are tracked via the audiobook_uploads junction table. When an audiobook is deleted, related chapters are automatically deleted via CASCADE, and upload_files are deleted via the chapters table foreign key relationship.';
