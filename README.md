# OpenShortPath

> ⚠️ **Warning: Active Development**
> 
> This repository is currently in **very active development**. APIs, configurations, and features may change without notice. Breaking changes are expected. Please use with caution in production environments.

OpenShortPath is a self-hostable URL shortening service with a modern web interface, comprehensive API, and flexible authentication options. Built with Go, React, and Next.js.

## 🏗️ Project Structure

OpenShortPath is organized into three main components:

```
OpenShortPath/
├── server/          # Go backend API server
├── dashboard/       # React frontend dashboard
├── landing/         # Next.js marketing/landing page
└── Makefile         # Build automation
```

## 🚀 Getting Started

### Prerequisites

- **Go 1.23+** (for server)
- **Node.js 18+** (for dashboard and landing)
- **PostgreSQL** (optional, SQLite is used by default)

### Development Setup

> 💡 **Recommended: Use Dev Containers**
> 
> For the best development experience, we recommend using [Dev Containers](https://containers.dev/) when possible. This ensures a consistent development environment across all contributors. If a `.devcontainer` configuration is available, open this repository in VS Code and select "Reopen in Container" when prompted.

#### Quick Start

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd OpenShortPath
   ```

2. **Set up the server**:
   ```bash
   cd server
   go mod download
   # Copy and configure config file
   cp config.example.yaml config.yaml
   # Edit config.yaml as needed
   go run . --config config.yaml
   ```

3. **Set up the dashboard** (optional, for development):
   ```bash
   cd dashboard
   npm install
   npm run dev
   ```

4. **Set up the landing page** (optional, for development):
   ```bash
   cd landing
   npm install
   npm run dev
   ```

#### Building for Production

Use the provided Makefile to build all components:

```bash
# Build server and dashboard
make server

# Clean build artifacts
make clean
```

This will:
1. Build the dashboard and copy it to `server/dashboard-dist/`
2. Build the Go server binary

The server embeds the dashboard, so you only need to run the server binary in production.

## 📚 Documentation

- **Server Documentation**: [`server/README.md`](server/README.md) - API setup, configuration, and features
- **Landing Documentation**: [`landing/README.md`](landing/README.md) - Landing page setup and customization

## 🛠️ Technology Stack

### Backend
- **Go 1.23+** - Programming language
- **Gin** - HTTP web framework
- **GORM** - ORM library
- **SQLite/PostgreSQL** - Database options
- **JWT** - Authentication

### Frontend
- **React 18** - UI library
- **TypeScript** - Type safety
- **Vite** - Build tool (dashboard)
- **Next.js 14** - React framework (landing)
- **Tailwind CSS** - Styling
- **React Router** - Client-side routing (dashboard)

## 🤝 Contributing

We welcome contributions! Please note that this project is in active development, so:

- APIs and features may change frequently
- Breaking changes are expected
- Documentation may lag behind implementation

When contributing:
1. Check existing issues and pull requests
2. Create a new issue to discuss major changes
3. Follow the existing code style and patterns
4. Add tests for new features
5. Update documentation as needed

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
