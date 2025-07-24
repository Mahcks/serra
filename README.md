# Serra 🎬

A modern, web-based media request and management system for your home media server.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Node Version](https://img.shields.io/badge/Node-18+-green.svg)](https://nodejs.org)

> ⚠️ **Alpha Software**: Serra is currently in alpha. Expect bugs and frequent changes. Use at your own risk.

## ✨ Features

- 🔍 **Smart Discovery**: Browse trending, popular, and upcoming content
- 📱 **Modern UI**: Clean, responsive interface built with React
- 👥 **User Management**: Role-based permissions and user accounts
- 🤖 **Auto-Approval**: Configurable automatic request approval
- 📺 **TV Show Support**: Season-specific requests with detailed tracking
- 🔄 **Real-time Updates**: Live status updates via WebSocket
- 🎯 **Advanced Requests**: On-behalf requests and bulk operations
- 📊 **Analytics**: Request trends and system monitoring
- 🔗 **Deep Integration**: Seamless connection with your media stack

## 🚀 Quick Start

### Prerequisites

- **Media Server**: Jellyfin or Emby
- **Arr Services**: Radarr (movies) and/or Sonarr (TV shows)
- **Download Client**: qBittorrent, SABnzbd, etc.

### Docker (Recommended)

```bash
git clone https://github.com/mahcks/serra
cd serra
docker-compose up -d
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
# Database
DATABASE_URL=./data/serra.db

# Security
JWT_SECRET=your-secret-key

# Media Server
MEDIA_SERVER_TYPE=jellyfin
MEDIA_SERVER_URL=http://localhost:8096
MEDIA_SERVER_API_KEY=your_api_key

# Services (optional)
RADARR_URL=http://localhost:7878
RADARR_API_KEY=your_radarr_key
SONARR_URL=http://localhost:8989
SONARR_API_KEY=your_sonarr_key
```

### Initial Setup

1. **Access Serra**: Open in your browser
2. **Create Admin**: Complete setup wizard
3. **Connect Services**: Link media server and arr services
4. **Configure Users**: Set up user accounts and permissions
5. **Test Requests**: Submit and approve test requests

## 🎯 Usage

### For Users
1. **Browse Content**: Explore trending and popular media
2. **Search**: Find specific movies and TV shows
3. **Request**: Click request buttons on content you want
4. **Track**: Monitor request status in "My Requests"
5. **Enjoy**: Watch approved content in your media server

### For Admins
1. **Manage Requests**: Review and approve user requests
2. **User Management**: Control permissions and access
3. **System Settings**: Configure services and preferences
4. **Monitor**: Track system health and usage analytics

## 🔐 Security

- JWT-based authentication
- Role-based permission system
- API rate limiting
- Input validation and sanitization
- Secure service integrations

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
go run ./cmd/app

# Frontend development  
cd frontend
npm run dev

# Database migrations
cd backend
go run ./cmd/app migrate
```

## 📊 Project Status

**Current Version**: Alpha  
**Stability**: Experimental  
**Production Ready**: No

### What Works
- ✅ User authentication and management
- ✅ Content discovery and search
- ✅ Request creation and approval
- ✅ Service integrations (Radarr/Sonarr)
- ✅ Real-time updates
- ✅ Basic admin features

### In Development
- 🚧 Mobile optimization
- 🚧 Advanced analytics
- 🚧 Plugin system
- 🚧 Multi-server support

### Known Issues
- Collection page status display
- Mobile layout optimization
- Performance with large libraries

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