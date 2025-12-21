# Incidex

<div align="center">

![Incidex Logo](./incidex_full_logo.jpg)

**Modern Incident Management System for SRE, DevOps, and Development Teams**

[English](./README_EN.md) | [日本語](./README.md)

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)
[![Next.js Version](https://img.shields.io/badge/Next.js-14+-000000?logo=next.js)](https://nextjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-3178C6?logo=typescript)](https://www.typescriptlang.org/)

</div>

---

## 📖 Overview

**Incidex** is an open-source incident management system that helps organizations record, manage, and learn from incidents through AI-powered summaries and postmortems.

By indexing incident information and accumulating organizational knowledge, Incidex helps prevent recurrence of similar incidents and supports team learning and continuous improvement.

### ✨ Key Features

- 🤖 **AI Summarization**: Automatically generate summaries from incident details (OpenAI API / Claude API support)
- 📊 **Timeline Management**: Record and visualize incident progression chronologically
- 🏷️ **Tag Management**: Flexible categorization and filtering with color-coded tags
- 📈 **Statistics Dashboard**: Visualize incident trends and track metrics like MTTR
- 📎 **File Attachments**: Manage related files such as logs and screenshots
- 🔍 **Advanced Search**: Fast search capabilities using PostgreSQL full-text search
- 📄 **PDF Report Generation**: Automatically generate summary reports for specified periods
- 🔐 **Self-Hosted**: Easy setup with Docker Compose, data stays within your organization
- 🌐 **Multi-Language Support**: UI support for Japanese and English (planned)

### 🎯 Target Users

- Small to medium-sized development teams and SRE teams (5-50 members)
- Security Operations Centers (SOC)
- IT departments and information systems departments
- Organizations prioritizing cost-effectiveness and self-hosting
- Organizations that cannot send data to external SaaS (financial institutions, government agencies, etc.)

---

## 🚀 Quick Start

### Prerequisites

- Docker 20.10+ and Docker Compose 2.0+
- Or, Go 1.21+ and Node.js 18+ installed

### Using Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/your-org/incidex.git
cd incidex

# Set up environment variables
cp .env.example .env
cp backend/.env.example backend/.env
cp frontend/.env.example frontend/.env.local

# For production environments, be sure to change the values in .env files
# See SECURITY.md for details

# Start the application
make up

# Or
docker-compose up -d
```

After starting, you can access the application at:

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **MinIO Console**: http://localhost:9090 (default: `minioadmin` / `minioadmin`)

### Local Development Setup

#### Backend

```bash
cd backend
cp .env.example .env
go mod download
go run cmd/server/main.go
```

#### Frontend

```bash
cd frontend
cp .env.example .env.local
npm install
npm run dev
```

For detailed setup instructions, see the [documentation](./docs/).

---

## 📋 Features

### Phase 1: Core Features (Implemented)

- ✅ **Authentication & User Management**
  - User registration and login (JWT authentication)
  - Role-based access control (Admin/Editor/Viewer)
  - Password hashing (bcrypt)

- ✅ **Incident Management**
  - Create, edit, delete, and list incidents
  - Severity (Critical/High/Medium/Low) and status management
  - Pagination, search, and filtering capabilities
  - SLA management and violation tracking

- ✅ **AI Summarization**
  - Automatic summary generation on incident creation
  - Manual summary regeneration
  - OpenAI API / Claude API support

- ✅ **Timeline Functionality**
  - Chronological event recording for incidents
  - Event types (detected, investigation started, root cause identified, mitigation, resolved, etc.)
  - Comment functionality

- ✅ **Tag Management**
  - Create, edit, and delete tags
  - Visual categorization with color settings
  - Filtering by tags

- ✅ **Dashboard**
  - Incident count trends (daily/weekly/monthly)
  - Distribution graphs by severity and status
  - Recent incidents list

- ✅ **File Attachments**
  - Attach files to incidents (images, PDFs, logs, etc.)
  - Object storage management with MinIO
  - File download and deletion

### Phase 2: Advanced Features (In Development)

- 🔄 **Postmortem Functionality**
  - Root cause analysis (Five Whys template)
  - Action item management
  - AI-assisted root cause analysis suggestions

- 🔄 **Advanced Search & Filtering**
  - PostgreSQL full-text search (Japanese and English support)
  - Multi-condition filtering
  - Redis caching of search results

- 🔄 **Statistics & Analytics**
  - MTTR (Mean Time To Recovery) calculation and display
  - Category-based incident trend analysis
  - Recurrence rate tracking

### Phase 3: Reporting Features (Planned)

- 📄 **PDF Generation**
  - PDF report output for individual incidents
  - Summary report generation for specified periods
  - Customizable report templates

---

## 🛠 Technology Stack

### Backend

- **Language**: Go 1.21+
- **Framework**: [Gin Web Framework](https://gin-gonic.com/)
- **ORM**: [GORM](https://gorm.io/)
- **Architecture**: Clean Architecture (`domain` / `usecase` / `interface` / `infrastructure`)
- **Database**: PostgreSQL 15+
- **Cache**: Redis 7+
- **Storage**: MinIO (S3-compatible object storage)
- **AI**: OpenAI API / Claude API

### Frontend

- **Framework**: [Next.js 14+](https://nextjs.org/) (App Router)
- **Language**: TypeScript 5+
- **Styling**: [TailwindCSS](https://tailwindcss.com/)
- **State Management**: React Context API

### Infrastructure

- **Containerization**: Docker & Docker Compose
- **Tooling**: Make (standardized development and startup commands)

---

## 📁 Project Structure

```
incidex/
├── backend/                 # Go Backend
│   ├── cmd/
│   │   ├── server/         # Main server
│   │   └── seed/          # Database seed tool
│   ├── internal/
│   │   ├── config/         # Configuration management
│   │   ├── domain/         # Domain entities & repository interfaces
│   │   ├── usecase/       # Business logic
│   │   ├── interface/     # HTTP handlers & routers
│   │   └── infrastructure/ # DB, storage, AI implementations
│   └── Dockerfile
├── frontend/               # Next.js Frontend
│   ├── src/
│   │   ├── app/           # App Router pages
│   │   ├── components/    # React components
│   │   ├── context/       # Global state management
│   │   ├── lib/           # API clients, etc.
│   │   └── types/         # TypeScript type definitions
│   └── Dockerfile
├── docs/                   # Documentation
│   ├── 要件定義書.md
│   ├── api-specification.md
│   ├── database-schema.md
│   └── プロジェクト計画書.md
├── docker-compose.yml      # Docker Compose configuration
├── Makefile               # Development commands
├── README.md              # This file (Japanese)
├── README_EN.md           # English README
├── SECURITY.md            # Security guidelines
├── CONTRIBUTING.md        # Contribution guide
└── LICENSE                # License file
```

---

## 📚 Documentation

Detailed documentation is available in the [`docs/`](./docs/) directory:

- [Requirements Specification](./docs/要件定義書.md) - Detailed functional and non-functional requirements
- [API Specification](./docs/api-specification.md) - Detailed REST API specifications
- [Database Schema](./docs/database-schema.md) - Database design
- [ER Diagram](./docs/er-diagram.md) - Entity relationship diagram
- [Project Plan](./docs/プロジェクト計画書.md) - Overall project plan

---

## 🔐 Security

Important security information is documented in [`SECURITY.md`](./SECURITY.md).

**Please read this before using in production.**

Key considerations:

- Strong `JWT_SECRET` configuration (minimum 32 characters)
- Database SSL enablement
- MinIO credential changes
- HTTPS/TLS configuration

If you discover a security vulnerability, please contact the project maintainers directly rather than opening a public issue.

---

## 🤝 Contributing

Contributions to Incidex are welcome!

See [`CONTRIBUTING.md`](./CONTRIBUTING.md) for detailed contribution guidelines.

### Types of Contributions

- 🐛 **Bug Reports**: Report issues via GitHub Issues
- 💡 **Feature Proposals**: Propose new features or improvements
- 🔧 **Code Improvements**: Improve code via Pull Requests
- 📝 **Documentation Improvements**: Improve or translate documentation
- 🧪 **Test Additions**: Increase test coverage

### Development Workflow

1. Fork this repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Create a Pull Request

---

## 📝 License

This project is licensed under the [MIT License](./LICENSE).

---

## 🗺 Roadmap

### Phase 1: MVP (Core Features) ✅
- Authentication & user management
- Incident CRUD, search, and filtering
- Tag management
- AI summarization
- Timeline functionality
- Dashboard

### Phase 2: Operational Enhancement 🔄
- Postmortem functionality
- Advanced search & filtering
- Extended statistics & analytics

### Phase 3: Reporting Features 📅
- PDF report generation
- Customizable report templates

### Future Plans
- Multi-tenant support (SaaS)
- Webhook notifications
- Slack integration
- Extended multi-language UI support
- Kubernetes Operator

For details, see the [Project Plan](./docs/プロジェクト計画書.md).

---

## 💬 Support

### Issue Reporting

Bug reports and feature requests are welcome via [GitHub Issues](https://github.com/your-org/incidex/issues).

### Discussions

General questions and discussions can be held in [GitHub Discussions](https://github.com/your-org/incidex/discussions).

### Security Issues

For security-related issues, please contact the project maintainers directly rather than opening a public issue.

---

## 🙏 Acknowledgments

Incidex depends on the following open-source projects:

- [Gin](https://gin-gonic.com/) - Go Web Framework
- [GORM](https://gorm.io/) - Go ORM
- [Next.js](https://nextjs.org/) - React Framework
- [TailwindCSS](https://tailwindcss.com/) - CSS Framework
- [PostgreSQL](https://www.postgresql.org/) - Database
- [Redis](https://redis.io/) - Cache
- [MinIO](https://min.io/) - Object Storage

Thanks to all developers of the dependency packages.

---

## 📞 Contact

- **GitHub**: [https://github.com/your-org/incidex](https://github.com/your-org/incidex)
- **Issues**: [https://github.com/your-org/incidex/issues](https://github.com/your-org/incidex/issues)
- **Discussions**: [https://github.com/your-org/incidex/discussions](https://github.com/your-org/incidex/discussions)

---

<div align="center">

**Made with ❤️ by the Incidex Team**

[⭐ Star us on GitHub](https://github.com/your-org/incidex) | [📖 Documentation](./docs/) | [🤝 Contribute](./CONTRIBUTING.md)

</div>

