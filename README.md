# Audio Book AI

A comprehensive AI-powered audiobook platform built with modern microservices architecture. This system allows users to upload audiobooks, automatically transcribe them, generate AI-powered summaries and tags, and provides a full-featured listening experience.

## 🏗️ Architecture Overview

The application uses a microservices architecture with the following components:

### Core Services

- **Next.js Web App** - Modern React-based frontend with TypeScript
- **Go API Service** - RESTful API server with JWT authentication
- **Go Worker** - Background AI processing (summarization, tagging, embeddings)
- **Go Transcriber** - Audio transcription service using Rev.ai
- **Redis** - Job queue for background processing
- **Supabase** - Cloud-based authentication, PostgreSQL database, and file storage

### External AI Services

- **Google Gemini API** - AI processing for summaries, tags, and embeddings
- **Rev.ai API** - Professional audio transcription service

## 📊 System Architecture Diagram

```mermaid
graph TB
    %% User Interfaces
    Web[Next.js Web App<br/>Port: 3000]

    %% Core Services
    API[Go API Service<br/>Port: 8080]
    Worker[Go Worker<br/>AI Processing]
    Transcriber[Go Transcriber<br/>Rev.ai Integration]
    Redis[(Redis<br/>Job Queue<br/>Port: 6379)]

    %% External Services
    Supabase[(Supabase<br/>Auth + PostgreSQL + Storage)]
    RevAI[Rev.ai API<br/>Audio Transcription]
    Gemini[Google Gemini API<br/>AI Processing]

    %% User Flow
    User((User)) --> Web
    Admin((Admin)) --> Web

    %% Web to API
    Web --> API

    %% API Dependencies
    API --> Supabase
    API --> Redis

    %% Worker Dependencies
    Worker --> Redis
    Worker --> Supabase
    Worker --> Gemini

    %% Transcriber Dependencies
    Transcriber --> Redis
    Transcriber --> Supabase
    Transcriber --> RevAI

    %% Job Flow
    API -.-> |"Queue Jobs"| Redis
    Redis -.-> |"Process Jobs"| Worker
    Redis -.-> |"Transcribe Audio"| Transcriber

    %% Data Flow
    API -.-> |"Store Files"| Supabase
    Worker -.-> |"AI Results"| Supabase
    Transcriber -.-> |"Transcripts"| Supabase

    %% Styling
    classDef service fill:#e1f5fe
    classDef external fill:#fff3e0
    classDef storage fill:#e8f5e8
    classDef user fill:#fce4ec

    class API,Worker,Transcriber,Web service
    class RevAI,Gemini external
    class Supabase,Redis storage
    class User,Admin user
```

## ✨ Features

### Admin Features

- ✅ Audiobook upload with drag & drop interface
- ✅ Cover image upload and management
- ✅ AI processing pipeline (transcription → summarization → tagging)
- ✅ Real-time job status monitoring
- ✅ Bulk audiobook management
- ✅ Price management and inventory control

### User Features

- ✅ Browse audiobook catalog with search and filters
- ✅ Personal library management
- ✅ Full-featured audio player with progress tracking
- ✅ Bookmark creation and management
- ✅ Shopping cart and purchase system
- ✅ Transcript and summary viewing
- ✅ Responsive mobile-friendly design

### Technical Features

- ✅ JWT-based authentication with Supabase
- ✅ Microservices architecture with Docker
- ✅ Background job processing with Redis
- ✅ File upload with progress tracking
- ✅ Real-time status updates
- ✅ Comprehensive error handling
- ✅ TypeScript support throughout

## 🚀 Quick Start

### Prerequisites

- **Docker & Docker Compose** - For containerized deployment
- **Supabase Account** - For authentication and database
- **Google Gemini API Key** - For AI processing
- **Rev.ai API Key** - For audio transcription

### System Requirements

- **RAM**: At least 4GB available
- **Storage**: 10GB available disk space
- **Ports**: 3000, 6379, 8080 should be available

### Environment Setup

1. **Clone the repository:**

   ```bash
   git clone <repository-url>
   cd audio-book-ai
   ```

2. **Set up environment files:**

   The project uses 5 separate environment files for different services. Copy the sample files and configure them:

   ```bash
   cd infra

   # Copy all sample environment files
   cp .common.env.sample .common.env
   cp .api.env.sample .api.env
   cp .worker.env.sample .worker.env
   cp .transcriber.env.sample .transcriber.env
   cp .web.env.sample .web.env
   ```

3. **Configure each environment file:**

   **`.common.env`** - Shared configuration:

   ```env
   INTERNAL_API_KEY=your_secure_random_key_here
   DATABASE_URL=postgresql://username:password@host:port/database?sslmode=require
   REDIS_URL=redis://redis:6379/0
   JOBS_PREFIX=audiobooks
   ```

   **`.api.env`** - API service configuration:

   ```env
   SUPABASE_URL=https://your-project-id.supabase.co
   SUPABASE_PUBLISHABLE_KEY=your_publishable_key_here
   SUPABASE_SECRET_KEY=your_secret_key_here
   SUPABASE_JWKS_URL=https://your-project-id.supabase.co/auth/v1/keys
   SUPABASE_STORAGE_BUCKET=audio
   ```

   **`.worker.env`** - AI processing worker:

   ```env
   GEMINI_API_KEY=your_gemini_api_key_here
   ```

   **`.transcriber.env`** - Audio transcription service:

   ```env
   REV_AI_API_KEY=your_rev_ai_api_key_here
   REV_AI_URL=https://api.rev.ai/speechtotext/v1
   ```

   **`.web.env`** - Frontend web application:

   ```env
   NEXT_PUBLIC_SUPABASE_URL=https://your-project-id.supabase.co
   NEXT_PUBLIC_SUPABASE_PUBLISHABLE_OR_ANON_KEY=your_publishable_key_here
   NEXT_PUBLIC_API_URL=http://localhost:8080
   ```

4. **Set up Supabase (one-time setup):**

   - Create a new Supabase project
   - Run the database schema from `api/database/schema.sql`
   - Configure storage bucket named "audio"
   - Set up authentication providers as needed

5. **Start all services:**

   ```bash
   docker-compose build
   docker-compose up -d
   ```

6. **Verify deployment:**
   - Web App: http://localhost:3000
   - API Health: http://localhost:8080/auth/health
   - Redis: Available on port 6379

**That's it!** The system is now running with all services containerized.

## 🛠️ Development Workflow

### Starting Development Environment

```bash
# Start all services (including web app)
cd infra
docker-compose up -d

# The web app is now available at http://localhost:3000
# All services are containerized and ready to use
```

### Service Management

```bash
# View all service status
docker-compose ps

# View logs
docker-compose logs -f [service-name]

# Restart a specific service
docker-compose restart [service-name]

# Stop all services
docker-compose down
```

### Building and Deploying

```bash
# Rebuild all services
docker-compose down
docker-compose build --no-cache
docker-compose up -d

# Build specific service
docker-compose build api
```

## 🔧 Key Architecture Decisions

### 1. Microservices Architecture

- **Decision**: Separate services for API, transcription, and AI processing
- **Rationale**: Better scalability, fault isolation, and technology flexibility
- **Trade-offs**: Increased complexity but better maintainability

### 2. Supabase for Backend Services

- **Decision**: Use Supabase for auth, database, and file storage
- **Rationale**: Reduces infrastructure complexity, provides robust auth
- **Trade-offs**: Vendor lock-in but significant development speed gains

### 3. Redis for Job Queue

- **Decision**: Use Redis for background job processing
- **Rationale**: Reliable job queue with good performance characteristics
- **Trade-offs**: Additional service dependency but enables async processing

### 4. Go for Backend Services

- **Decision**: Use Go for all backend microservices
- **Rationale**: Excellent performance, strong typing, great concurrency support
- **Trade-offs**: Learning curve but better performance than alternatives

### 5. Next.js for Frontend

- **Decision**: Use Next.js with TypeScript and Tailwind CSS
- **Rationale**: Modern React framework with excellent developer experience
- **Trade-offs**: Framework complexity but rapid development capabilities

### 6. JWT Authentication Strategy

- **Decision**: Use Supabase JWT tokens with JWKS validation
- **Rationale**: Stateless authentication with proper security practices
- **Trade-offs**: Token management complexity but better scalability

## 📋 API Endpoints

### Authentication Routes

- `POST /auth/validate` - Validate JWT token
- `GET /auth/me` - Get current user information
- `GET /auth/health` - Health check endpoint

### User Routes (Protected)

- `GET /user/audiobooks` - Get user's audiobooks
- `POST /user/cart` - Add audiobook to cart
- `GET /user/cart` - Get user's cart
- `POST /user/checkout` - Process purchase
- `GET /user/purchases` - Get purchase history

### Admin Routes (Admin Only)

- `POST /admin/audiobooks` - Create new audiobook
- `GET /admin/uploads/:id/progress` - Get upload progress
- `POST /admin/audiobooks/:id/trigger-summarize-tag` - Trigger AI processing
- `GET /admin/audiobooks/:id/jobs` - Get processing job status

### Public Routes

- `GET /public/audiobooks` - Browse public audiobooks
- `GET /public/audiobooks/:id` - Get specific audiobook details

## 🧪 Test Coverage

The project includes comprehensive test coverage across all API endpoints:

### Test Statistics

- **Total Test Files**: 6 (auth, cart, admin, jobs, internal, audiobooks)
- **Coverage Areas**: All HTTP methods (GET, POST, PUT, DELETE)
- **Test Scenarios**: Success cases, validation errors, auth errors, database errors
- **Mock Implementation**: Complete repository mocking for isolated testing

### Test Files Overview

| Test File            | Coverage                 | Test Count |
| -------------------- | ------------------------ | ---------- |
| `auth_test.go`       | Authentication routes    | 15+ tests  |
| `cart_test.go`       | Cart and purchase routes | 25+ tests  |
| `admin_test.go`      | Admin management routes  | 30+ tests  |
| `jobs_test.go`       | Job management routes    | 20+ tests  |
| `internal_test.go`   | Internal API routes      | 15+ tests  |
| `audiobooks_test.go` | Core audiobook routes    | 35+ tests  |

### Running Tests

```bash
# Run all tests
cd api
go test ./test/... -v

# Run tests with coverage
go test ./test/... -v -cover

# Generate coverage report
go test ./test/... -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Results Summary

- ✅ **140+ individual test cases** covering all major functionality
- ✅ **Complete mock implementation** for database operations
- ✅ **Error scenario coverage** for robust error handling
- ✅ **Authentication testing** for security validation
- ✅ **HTTP status code verification** for proper API responses

## 📊 Performance Considerations

### Scalability Features

- **Horizontal scaling**: Stateless services can be scaled independently
- **Job queue**: Background processing prevents blocking operations
- **File streaming**: Efficient audio file handling for large files
- **Database optimization**: Proper indexing and query optimization

### Resource Usage

- **API Service**: ~50MB RAM per instance
- **Worker Service**: ~100MB RAM per instance (AI processing)
- **Transcriber Service**: ~75MB RAM per instance
- **Redis**: ~50MB RAM for job queue
- **Total**: ~275MB minimum RAM requirement

## 🔒 Security Features

### Authentication & Authorization

- JWT-based authentication with Supabase
- Role-based access control (Admin vs User)
- API key protection for internal services
- CORS configuration for frontend integration

### Data Protection

- Encrypted file storage in Supabase
- Secure API key management
- Input validation and sanitization
- SQL injection prevention with parameterized queries

## 📁 Project Structure

```
audio-book-ai/
├── api/                    # Go API service
│   ├── handlers/          # HTTP handlers
│   ├── middleware/        # Authentication middleware
│   ├── models/           # Data models
│   ├── database/         # Database layer
│   ├── services/         # External service integrations
│   └── test/            # Comprehensive test suite
├── web/                   # Next.js frontend
│   ├── app/              # App router pages
│   ├── components/       # React components
│   ├── hooks/           # Custom React hooks
│   └── lib/             # Utility libraries
├── worker/               # Go AI processing worker
├── transcriber/         # Go transcription worker
└── infra/               # Docker infrastructure
    ├── docker-compose.yml
    └── env.example
```

## 🚧 Deployment Options

### Development

- Local Docker Compose setup
- Hot reloading for web development
- Separate service containers for debugging

### Production

- Docker Swarm or Kubernetes deployment
- Environment-specific configuration
- Load balancing and auto-scaling
- Monitoring and logging integration

## 📖 Documentation

- [API Documentation](api/docs/README.md)
- [Web App Documentation](web/docs/README.md)
- [Infrastructure Guide](infra/docs/README.md)
- [Complete Audiobook Flow](api/docs/COMPLETE_AUDIOBOOK_FLOW.md)
- [Test Documentation](api/docs/COMPREHENSIVE_TEST_SUMMARY.md)
- [Storage Setup Guide](infra/docs/STORAGE_SETUP.md)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Run tests (`go test ./test/... -v`)
4. Commit your changes (`git commit -m 'Add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For issues and questions:

1. Check the documentation in each service's `docs/` directory
2. Review the troubleshooting section in `infra/docs/README.md`
3. Check existing GitHub issues
4. Create a new issue with detailed information

## 🔮 Future Enhancements

- [ ] Real-time collaboration features
- [ ] Cover image upload
- [ ] Advanced analytics dashboard
- [ ] Mobile app development
- [ ] Integration with more AI providers
- [ ] Advanced search with vector embeddings
- [ ] Multi-language support
- [ ] Performance monitoring dashboard
