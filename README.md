# Helga

A powerful tool for synchronizing Helm packages from JFrog Artifactory to Kubernetes clusters automatically.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

## Overview

Helga is a Kubernetes-native tool designed to automate the synchronization of Helm packages from JFrog Artifactory repositories to your Kubernetes clusters. It continuously monitors specified Artifactory repositories and automatically deploys or updates Helm charts based on configurable policies.

### Key Benefits

- **Automated Deployment**: Automatically sync Helm packages from Artifactory to your clusters
- **Multi-Cluster Support**: Manage multiple Kubernetes clusters from a single configuration
- **Flexible Versioning**: Choose between version-based or timestamp-based package selection
- **Concurrent Processing**: Efficient concurrent synchronization across namespaces
- **Robust Error Handling**: Comprehensive error handling and logging

## Features

- 🔄 **Continuous Synchronization**: Automatically sync Helm packages at configurable intervals
- 🎯 **Multi-Target Support**: Deploy to multiple clusters and namespaces simultaneously
- 📦 **Artifactory Integration**: Native integration with JFrog Artifactory using AQL queries
- 🔀 **Version Strategy**: Support for both semantic versioning and timestamp-based selection
- 🛡️ **Security**: Support for both token-based and password-based authentication
- 📊 **Comprehensive Logging**: Detailed logging with structured JSON and console output
- ⚡ **Performance**: Concurrent processing with caching mechanisms

## Architecture

Helga is built with a modular architecture:

```
helga/
├── cmd/main/           # Application entry point
├── internal/
│   ├── logger/         # Structured logging
│   ├── utils/          # Utility functions
│   └── vars/           # Environment variables and constants
└── pkg/
    ├── config/         # Configuration management
    ├── errors/         # Custom error types
    └── models/         # Core business logic
        ├── artifact.go    # Artifactory integration
        ├── cluster.go     # Kubernetes cluster management
        ├── helm.go        # Helm chart handling
        ├── namespace.go   # Namespace-level operations
        └── repo.go        # Repository management
```

### Core Components

- **`config.Config`**: Manages application configuration and validation
- **`models.Cluster`**: Handles Kubernetes cluster connections and operations
- **`models.Artifact`**: Manages JFrog Artifactory integration
- **`models.Namespace`**: Orchestrates namespace-level synchronization
- **`models.Repo`**: Handles repository configuration and validation

## Installation

### Prerequisites

- Go 1.24.3 or later
- Access to a JFrog Artifactory instance
- Kubernetes cluster access with appropriate permissions
- `golangci-lint` (for development)

### From Source

```bash
# Clone the repository
git clone <repository-url>
cd helga

# Build the application
make build

# The binary will be created at ./bin/helga
```

### Using Make Targets

```bash
# Build and run in one command
make run

# Or build, test, and run with full quality checks
make all
```

## Configuration

Helga uses a YAML configuration file to define clusters, namespaces, and Artifactory settings.

### Environment Variables

Set the following environment variables:

```bash
export HELGA_CONF_FILE_PATH="/path/to/your/config.yaml"
export LOGS_FILE_PATH="/path/to/helga.log"
```

### Configuration Structure

The configuration consists of two main sections:

#### Global Configuration

Defines default settings that can be inherited by individual clusters:

```yaml
global:
  cluster:
    insecure_skip_tls_verify: true
    # ca_cert_file_path: "/path/to/ca.crt"  # Required if insecure_skip_tls_verify is false
  artifact:
    domain: "https://your-artifactory.com/artifactory"
    username: "your-username"
    password: "your-password"
    repos:
      - name: "helm-repo"
        decideByVersion: false  # Use timestamp if false, semantic versioning if true
        paths:
          - "/path/to/charts"
```

#### Cluster Configuration

Defines specific clusters and their namespaces:

```yaml
clusters:
  - name: "production-cluster"
    server: "https://k8s-api.example.com"
    username: "cluster-user"
    password: "cluster-password"  # Or use token instead
    # token: "your-k8s-token"
    insecure_skip_tls_verify: true
    namespaces:
      - name: "webapp"
        sync_interval: 300  # Sync every 5 minutes
        artifact:
          repos:
            - name: "helm-repo"
              paths:
                - "/webapp/charts"
```

### Complete Example

See `helga_conf_example.yaml` for a complete configuration example.

## Usage

### Running Helga

```bash
# Set environment variables
export HELGA_CONF_FILE_PATH="./helga_conf_example.yaml"
export LOGS_FILE_PATH="./helga.log"

# Run the application
make run
```

### Configuration Validation

Helga performs comprehensive validation of your configuration:

- **Cluster connectivity**: Validates Kubernetes API server access
- **Artifactory access**: Verifies repository permissions and connectivity  
- **Helm repository validation**: Ensures chart repositories are accessible
- **Sync interval validation**: Enforces minimum sync intervals for stability

### Synchronization Process

1. **Initialization**: Helga connects to all configured clusters and validates access
2. **Repository Setup**: Adds configured Helm repositories to each namespace
3. **Continuous Sync**: For each namespace:
   - Queries Artifactory for available charts
   - Compares with deployed releases
   - Determines updates needed
   - Deploys or upgrades charts as necessary

### Version Selection Strategy

Helga supports two strategies for selecting chart versions:

- **Semantic Versioning** (`decideByVersion: true`): Uses semantic version comparison
- **Timestamp-based** (`decideByVersion: false`): Uses the most recently modified chart

## Development

### Make Targets

The `Makefile` provides several useful targets:

```bash
# Development workflow
make setup          # Tidy Go modules
make check-quality   # Run linting, formatting, and vet
make test-coverage   # Run tests with coverage report
make build          # Build the application
make run            # Build and run
make clean          # Clean generated files

# Code quality
make lint           # Run golangci-lint
make fmt            # Format code
make vet            # Run go vet
make test           # Run tests

# Playground (for experimentation)
make pg-init        # Create playground directory
make pg-build       # Build playground
make pg-run         # Run playground
make pg-clean       # Clean playground
```

### Project Structure

- **`cmd/main/main.go`**: Application entry point with graceful shutdown
- **`internal/logger`**: Structured logging with file and console output
- **`internal/vars/`**: Environment variables and constants
- **`pkg/config`**: Configuration loading and validation
- **`pkg/errors`**: Custom error types with context
- **`pkg/models`**: Core business logic and data models

### Error Handling

Helga uses a comprehensive error handling system with custom error types for different components:

- Configuration validation errors
- Artifactory API errors
- Helm client errors
- Centralized error logging

### Testing

```bash
# Run tests with coverage
make test-coverage

# Run only tests
make test
```

### Code Quality

The project uses `golangci-lint` with strict quality rules:

```bash
# Check code quality
make check-quality

# Individual checks
make lint    # Linting
make fmt     # Formatting  
make vet     # Static analysis
```

### Playground

For experimentation and testing:

```bash
# Initialize playground
make pg-init

# Build and run playground
make pg-run

# Clean playground
make pg-clean
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Ensure tests pass (`make test`)
5. Ensure code quality (`make check-quality`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Development Guidelines

- Follow Go best practices and idioms
- Maintain test coverage above 80%
- Ensure all linting rules pass
- Add comprehensive documentation for new features
- Use structured logging consistently

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Helga** - Keeping your Kubernetes clusters in sync with your artifact repositories, automatically and reliably.