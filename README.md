# Serra 🎬

A modern, web-based media request and management system for your home media server.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Node Version](https://img.shields.io/badge/Node-18+-green.svg)](https://nodejs.org)

> 🚨 **ALPHA SOFTWARE - NOT FOR PRODUCTION** 🚨  
> Serra is in early alpha development. **Breaking changes are frequent and expected.** Database schemas, API endpoints, and configuration formats **WILL change** without notice. **Do not use in production environments.** Data loss is possible during updates. This software is intended for developers and early testers only.

## ✨ Features

- 🔍 **Smart Discovery**: Browse trending, popular, and upcoming content with TMDB integration
- 📱 **Modern UI**: Clean, responsive interface built with React and Tailwind CSS
- 👥 **User Management**: Role-based permissions, user invitations, and account management
- 🎫 **Invitation System**: Create shareable invitation links with optional email delivery
- 🤖 **Auto-Approval**: Configurable automatic request approval based on user permissions
- 📺 **TV Show Support**: Season-specific requests with detailed episode tracking
- 🔄 **Real-time Updates**: Live status updates via WebSocket connections
- 🎯 **Advanced Requests**: On-behalf requests and bulk operations for admins
- 📊 **Analytics Dashboard**: Request trends, storage monitoring, and system health metrics
- 🗄️ **Storage Management**: Monitor drive usage, set thresholds, and track storage pools
- 🔗 **Deep Integration**: Seamless connection with your entire media stack

## 🚀 Quick Start

### Prerequisites

- **Media Server**: Jellyfin or Emby
- **Arr Services**: Radarr (movies) and/or Sonarr (TV shows)
- **Download Client**: qBittorrent, SABnzbd, etc.

### Docker (Recommended)

Serra is distributed as two separate containers via GitHub Container Registry:

```bash
# Backend container
docker run -d \
  --name serra-backend \
  -p 9090:9090 \
  -v $(pwd)/data:/app/data \
  -e SQLITE_PATH=/app/data/serra.db \
  -e CREDENTIALS_JWT_SECRET=your-secret-key \
  ghcr.io/mahcks/serra-backend:latest

# Frontend container  
docker run -d \
  --name serra-frontend \
  -p 3000:3000 \
  -e VITE_API_BASE_URL=http://localhost:9090/v1 \
  ghcr.io/mahcks/serra-frontend:latest

# Custom ports example
# Backend on port 8080, frontend on port 8081
docker run -d --name serra-backend -p 8080:9090 ... ghcr.io/mahcks/serra-backend:latest
docker run -d --name serra-frontend -p 8081:3000 -e VITE_API_BASE_URL=http://localhost:8080/v1 ... ghcr.io/mahcks/serra-frontend:latest
```

Or use Docker Compose:

```yaml
version: '3.8'
services:
  serra-backend:
    image: ghcr.io/mahcks/serra-backend:latest
    ports:
      - "9090:9090"    # Change left port for custom host port
    volumes:
      - ./data:/app/data
    environment:
      - SQLITE_PATH=/app/data/serra.db
      - CREDENTIALS_JWT_SECRET=your-secret-key

  serra-frontend:
    image: ghcr.io/mahcks/serra-frontend:latest
    ports:
      - "3000:3000"    # Change left port for custom host port
    environment:
      - VITE_API_BASE_URL=http://localhost:9090/v1
    depends_on:
      - serra-backend

  # Example: Custom ports
  # serra-backend:
  #   ports:
  #     - "8080:9090"  # Backend accessible on host port 8080
  # serra-frontend:
  #   ports:
  #     - "8081:3000"  # Frontend accessible on host port 8081
  #   environment:
  #     - VITE_API_BASE_URL=http://localhost:8080/v1
```

Visit `http://localhost:3000` and complete the setup wizard.

### Manual Installation

**Backend**:
```bash
cd backend
go mod tidy
go build -o serra ./cmd/app
./serra
```

**Frontend**:
```bash
cd frontend
npm install
npm run build
npm run preview
```

## 📖 Documentation

- **[Installation Guide](docs/INSTALLATION.md)**: Detailed setup instructions
- **[User Guide](docs/USER_GUIDE.md)**: How to use Serra as an end user
- **[Admin Guide](docs/ADMIN_GUIDE.md)**: Administration and configuration
- **[API Documentation](docs/API.md)**: REST API reference
- **[Troubleshooting](docs/TROUBLESHOOTING.md)**: Common issues and solutions
- **[Known Issues](docs/KNOWN_ISSUES.md)**: Current limitations and workarounds

## 🛠️ Technology Stack

**Backend**:
- Go 1.21+ with Fiber web framework
- SQLite database with migrations
- JWT authentication
- WebSocket real-time updates

**Frontend**:
- React 18 with TypeScript
- Tailwind CSS for styling
- React Query for state management
- Radix UI components

**Integrations**:
- TMDB API for media metadata
- Radarr/Sonarr for download automation
- Jellyfin/Emby for library management
- qBittorrent/SABnzbd for downloads

## 🔧 Configuration

### Environment Variables

```bash
# Backend Configuration
REST_ADDRESS=0.0.0.0        # Server bind address
REST_PORT=9090               # Server port (can be overridden for Docker)
SQLITE_PATH=./data/serra.db  # Database file path
CREDENTIALS_JWT_SECRET=your-secret-key

# Frontend Configuration  
VITE_API_BASE_URL=http://localhost:9090/v1

# Media Server (configured via web interface)
MEDIA_SERVER_TYPE=jellyfin
MEDIA_SERVER_URL=http://localhost:8096
MEDIA_SERVER_API_KEY=your_api_key

# Services (configured via web interface)
RADARR_URL=http://localhost:7878
RADARR_API_KEY=your_radarr_key
SONARR_URL=http://localhost:8989
SONARR_API_KEY=your_sonarr_key

# Email (optional - for invitation delivery)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_app_password
```

### Initial Setup

1. **Access Serra**: Open `http://localhost:3000` in your browser
2. **Setup Wizard**: Complete initial configuration (admin account, media server)
3. **Connect Services**: Link Radarr, Sonarr, and download clients
4. **Configure Users**: Set default permissions and create invitations
5. **Email Setup** (Optional): Configure SMTP for invitation emails
6. **Test Requests**: Submit and approve test requests to verify workflow

## 🎯 Usage

### For Users
1. **Browse Content**: Explore trending and popular media
2. **Search**: Find specific movies and TV shows
3. **Request**: Click request buttons on content you want
4. **Track**: Monitor request status in "My Requests"
5. **Enjoy**: Watch approved content in your media server

### For Admins
1. **Manage Requests**: Review and approve user requests with bulk operations
2. **User Management**: Create invitations, manage permissions, and control access
3. **System Settings**: Configure services, email delivery, and default permissions
4. **Monitor**: Track system health, storage usage, and analytics dashboard
5. **Invitations**: Send invitation links directly or via email to new users

## 🔐 Security

- JWT-based authentication with refresh tokens
- Role-based permission system with granular controls
- CSRF protection on all state-changing operations
- API rate limiting and invitation throttling
- Input validation and sanitization
- Secure service integrations with API key management
- Background cleanup jobs for expired invitations
- Protected admin routes and middleware

**🔒 Security Requirements**:
- **Change JWT secret** - Set `CREDENTIALS_JWT_SECRET` environment variable
- **Use HTTPS** in production for secure token transmission
- **Keep dependencies updated** for security patches
- **Regular backups** of user data and configurations

## 🤝 Contributing

Serra is open source and welcomes contributions!

1. **Fork** the repository
2. **Create** a feature branch
3. **Make** your changes
4. **Add** tests if applicable
5. **Submit** a pull request

### Development Setup

```bash
# Backend development
cd backend
make run

# Frontend development  
cd frontend
npm run dev

# Database migrations
cd backend
make migrate

# Generate TypeScript types
cd backend
make tygo

# Switch between media servers (development only)
./scripts/switch-media-server.sh jellyfin  # Switch to Jellyfin
./scripts/switch-media-server.sh emby      # Switch to Emby  
./scripts/switch-media-server.sh status    # Check current setup
```

The `switch-media-server.sh` script helps developers easily switch between Jellyfin and Emby databases during development and testing.

## 📊 Project Status

**Current Version**: Alpha  
**Stability**: Experimental  
**Production Ready**: No

### What Works
- ✅ User authentication and management with invitation system
- ✅ Content discovery and search with TMDB integration
- ✅ Request creation, approval, and lifecycle management
- ✅ Service integrations (Radarr/Sonarr/Download clients)
- ✅ Real-time updates and notifications
- ✅ Admin dashboard with analytics and monitoring
- ✅ Storage management and drive monitoring
- ✅ Email system for notifications (optional)
- ✅ CSRF protection and rate limiting
- ✅ User permission management and role-based access

### In Development
- 🚧 Mobile app companion
- 🚧 Advanced analytics and reporting
- 🚧 Plugin system for extensibility
- 🚧 Multi-server support
- 🚧 Advanced search filters
- 🚧 Custom themes and branding

### Known Issues
- Collection page status display needs refinement
- Mobile layout optimization in progress
- Performance with very large libraries (10k+ items)

See [Known Issues](docs/KNOWN_ISSUES.md) for complete list.

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **TMDB**: Media metadata and images
- **Radarr/Sonarr**: Download automation
- **Jellyfin/Emby**: Media server platforms
- **React/Go**: Core technologies
- **Community**: Testing and feedback

## 📞 Support

- **Documentation**: Check the [docs](docs/) directory
- **GitHub Issues**: Report bugs and request features
- **Discord**: Join the community for help and discussion
- **Troubleshooting**: See the [troubleshooting guide](docs/TROUBLESHOOTING.md)

---

**Built with ❤️ for the self-hosted media community**

[![GitHub stars](https://img.shields.io/github/stars/mahcks/serra?style=social)](https://github.com/mahcks/serra/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/mahcks/serra?style=social)](https://github.com/mahcks/serra/network/members)
[![GitHub issues](https://img.shields.io/github/issues/mahcks/serra)](https://github.com/mahcks/serra/issues)