<p align="center">
  <pre align="center">
          ,_---~~~~~----._         
   _,,_,*^____      _____``*g*\"*, 
  / __/ /'     ^.  /      \ ^@q   f 
 [  @f | @))    |  | @))   l  0 _/  
  \`/   \~____ / __ \_____/    \   
   |           _l__l_           I   
   }          [______]           I  
   ]            | | |            |  
   ]             ~ ~             |  
   |                            |   
    |                           |   
  </pre>
</p>

<p align="center">
  A powerful <a href="https://go.dev" target="_blank">Go</a> project scaffolding tool for generating production-ready service templates.
</p>

<p align="center">
  <a href="#"><img src="https://img.shields.io/badge/go-1.23+-00ADD8?style=flat&logo=go" alt="Go Version" /></a>
  <a href="#"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License" /></a>
  <a href="#"><img src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg" alt="PRs Welcome" /></a>
</p>

## Description

GoG (Go Generator) is a CLI tool that helps you quickly scaffold production-ready Go services. It generates a well-structured project template with best practices and common integrations pre-configured.

## Generated Project Features

- 📦 **DDD Structure** - Domain-driven design project layout
- 🚀 **Fiber Integration** - High-performance HTTP server
- 🔄 **Database Ready** - PostgreSQL setup with migrations
- 📨 **NATS Ready** - Message queue integration
- 🔐 **Security** - Pre-configured authentication and authorization
- 📝 **API Docs** - Swagger/OpenAPI documentation
- 🐳 **Docker Ready** - Containerization setup
- ⚡ **Development Tools** - Migration and other utilities included

## Installation

```bash
# Install GoG CLI
go install github.com/mohamedalosaili/gog/cmd/gog@latest
```

## Usage

```bash
# Create a new project
gog new new-service -u github_username

# Generate a new migration
just migrate-new create_users_table
# or 
just mn create_users_table
```

## Generated Project Structure

```
.
├── cmd/                    # Application entry points
│   ├── migrate/           # Database migrations commands
│   └── serve/             # HTTP server
├── internal/              # Private application code
│   ├── config/           # Configuration
│   ├── domains/          # Business logic
│   │   ├── health/      # Health check domain
│   │   ├── post/        # Post domain example
│   │   └── user/        # User domain example
│   ├── middleware/      # HTTP middleware
│   └── registry/        # Dependency injection
└── migrations/           # Database migrations
```

## Template Configuration

The generated project includes a configuration file:

```yaml
# config.yaml
app:
  name: "[YOUR_APP_NAME]"
  env: "development"
  port: 3000

database:
  host: "localhost"
  port: 5432
  name: "mydb"
```

## Available Commands in Generated Project

```bash
# Development
go run main.go serve              # Start the server
go run main.go migrate up         # Run migrations
go run main.go migrate down       # Rollback migrations
go run main.go migrate status     # Check migration status
```

## Post-Generation Steps

After generating your project:

1. Update project configuration:
```bash
cp config.yaml.example config.yaml
```


3. Start development:
```bash
# Start infrastructure
docker-compose up -d

# Run your service
go run main.go serve
# or
just serve
```

## Generated API Documentation

Your generated service will include Swagger documentation at:
```
http://localhost:3000/api/docs
```

## Template Features

The generated template includes:

- ✅ Structured logging
- ✅ Configuration management
- ✅ Database migrations
- ✅ Health checks
- ✅ API documentation
- ✅ Error handling
- ✅ Dependency injection
- ✅ Docker support
- ✅ Example domains

## Support

GoG is an MIT-licensed open source project. It can grow thanks to sponsors and support from the community.


## Contributing

We welcome contributions to improve the scaffolding templates!

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing`)
5. Open a Pull Request

## License

GoG is [MIT licensed](LICENSE).

---