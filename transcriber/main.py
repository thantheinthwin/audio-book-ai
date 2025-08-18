#!/usr/bin/env python3
"""
Audio Book AI - Transcriber Worker

This worker processes audio files using OpenAI Whisper and saves transcripts to the database.
"""

import os
import json
import time
import logging
import whisper
import redis
import psycopg2
from psycopg2.extras import RealDictCursor
from dotenv import load_dotenv
import uuid
from datetime import datetime

# Load environment variables
load_dotenv()

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class TranscriberWorker:
    def __init__(self):
        # Redis connection
        self.redis_url = os.getenv('REDIS_URL', 'redis://redis:6379/0')
        self.jobs_prefix = os.getenv('JOBS_PREFIX', 'audiobooks')
        
        # Database connection
        self.db_url = os.getenv('DATABASE_URL')
        
        # Whisper configuration
        self.whisper_model = os.getenv('WHISPER_MODEL', 'tiny')
        self.whisper_language = os.getenv('WHISPER_LANGUAGE', 'auto')
        
        # Initialize connections
        self.redis_client = None
        self.db_conn = None
        self.whisper_model_instance = None
        
    def connect_redis(self):
        """Connect to Redis"""
        try:
            self.redis_client = redis.from_url(self.redis_url)
            self.redis_client.ping()
            logger.info("Connected to Redis")
        except Exception as e:
            logger.error(f"Failed to connect to Redis: {e}")
            raise
    
    def connect_database(self):
        """Connect to PostgreSQL database"""
        try:
            self.db_conn = psycopg2.connect(self.db_url)
            logger.info("Connected to database")
        except Exception as e:
            logger.error(f"Failed to connect to database: {e}")
            raise
    
    def load_whisper_model(self):
        """Load Whisper model"""
        try:
            logger.info(f"Loading Whisper model: {self.whisper_model}")
            self.whisper_model_instance = whisper.load_model(self.whisper_model)
            logger.info("Whisper model loaded successfully")
        except Exception as e:
            logger.error(f"Failed to load Whisper model: {e}")
            raise
    
    def get_pending_jobs(self):
        """Get pending transcription jobs from database"""
        try:
            with self.db_conn.cursor(cursor_factory=RealDictCursor) as cursor:
                cursor.execute("""
                    SELECT pj.*, ab.file_path, ab.language
                    FROM processing_jobs pj
                    JOIN audiobooks ab ON pj.audiobook_id = ab.id
                    WHERE pj.job_type = 'transcribe' 
                    AND pj.status = 'pending'
                    ORDER BY pj.created_at ASC
                    LIMIT 5
                """)
                return cursor.fetchall()
        except Exception as e:
            logger.error(f"Failed to get pending jobs: {e}")
            return []
    
    def update_job_status(self, job_id, status, error_message=None):
        """Update job status in database"""
        try:
            with self.db_conn.cursor() as cursor:
                if status == 'running':
                    cursor.execute("""
                        UPDATE processing_jobs 
                        SET status = %s, started_at = NOW()
                        WHERE id = %s
                    """, (status, job_id))
                elif status in ['completed', 'failed']:
                    cursor.execute("""
                        UPDATE processing_jobs 
                        SET status = %s, completed_at = NOW(), error_message = %s
                        WHERE id = %s
                    """, (status, error_message, job_id))
                
                self.db_conn.commit()
                logger.info(f"Updated job {job_id} status to {status}")
        except Exception as e:
            logger.error(f"Failed to update job status: {e}")
            self.db_conn.rollback()
    
    def transcribe_audio(self, audio_path):
        """Transcribe audio file using Whisper"""
        try:
            start_time = time.time()
            
            # Transcribe with Whisper
            result = self.whisper_model_instance.transcribe(
                audio_path,
                language=self.whisper_language,
                verbose=False
            )
            
            processing_time = int(time.time() - start_time)
            
            return {
                'content': result['text'],
                'segments': result.get('segments', []),
                'language': result.get('language', 'unknown'),
                'processing_time_seconds': processing_time
            }
        except Exception as e:
            logger.error(f"Failed to transcribe audio: {e}")
            raise
    
    def save_transcript(self, audiobook_id, transcript_data):
        """Save transcript to database"""
        try:
            with self.db_conn.cursor() as cursor:
                cursor.execute("""
                    INSERT INTO transcripts (
                        id, audiobook_id, content, segments, language, 
                        confidence_score, processing_time_seconds, created_at
                    ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
                    ON CONFLICT (audiobook_id) DO UPDATE SET
                        content = EXCLUDED.content,
                        segments = EXCLUDED.segments,
                        language = EXCLUDED.language,
                        confidence_score = EXCLUDED.confidence_score,
                        processing_time_seconds = EXCLUDED.processing_time_seconds
                """, (
                    str(uuid.uuid4()),
                    str(audiobook_id),
                    transcript_data['content'],
                    json.dumps(transcript_data['segments']) if transcript_data['segments'] else None,
                    transcript_data['language'],
                    0.95,  # Default confidence score
                    transcript_data['processing_time_seconds'],
                    datetime.now()
                ))
                
                self.db_conn.commit()
                logger.info(f"Saved transcript for audiobook {audiobook_id}")
        except Exception as e:
            logger.error(f"Failed to save transcript: {e}")
            self.db_conn.rollback()
            raise
    
    def process_job(self, job):
        """Process a single transcription job"""
        job_id = job['id']
        audiobook_id = job['audiobook_id']
        file_path = job['file_path']
        
        logger.info(f"Processing transcription job {job_id} for audiobook {audiobook_id}")
        
        try:
            # Update job status to running
            self.update_job_status(job_id, 'running')
            
            # Check if file exists
            if not os.path.exists(file_path):
                raise FileNotFoundError(f"Audio file not found: {file_path}")
            
            # Transcribe audio
            transcript_data = self.transcribe_audio(file_path)
            
            # Save transcript to database
            self.save_transcript(audiobook_id, transcript_data)
            
            # Update job status to completed
            self.update_job_status(job_id, 'completed')
            
            logger.info(f"Successfully processed transcription job {job_id}")
            
        except Exception as e:
            error_msg = str(e)
            logger.error(f"Failed to process job {job_id}: {error_msg}")
            self.update_job_status(job_id, 'failed', error_msg)
    
    def run(self):
        """Main worker loop"""
        logger.info("Starting Transcriber Worker")
        
        # Initialize connections and model
        self.connect_redis()
        self.connect_database()
        self.load_whisper_model()
        
        logger.info("Transcriber Worker ready to process jobs")
        
        while True:
            try:
                # Get pending jobs
                pending_jobs = self.get_pending_jobs()
                
                if pending_jobs:
                    logger.info(f"Found {len(pending_jobs)} pending transcription jobs")
                    
                    for job in pending_jobs:
                        self.process_job(job)
                else:
                    # No jobs, wait a bit
                    time.sleep(10)
                    
            except Exception as e:
                logger.error(f"Error in main loop: {e}")
                time.sleep(30)  # Wait longer on error
                
            # Small delay between iterations
            time.sleep(5)

if __name__ == "__main__":
    worker = TranscriberWorker()
    worker.run()
