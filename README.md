# HHX - Headless Hawx CLI

HHX (Headless Hawx) is a command-line tool for efficiently managing database tables and storage resources. It provides
an intuitive workflow for tracking changes, staging files, and synchronizing content with remote servers.

Designed for developers and data professionals who need streamlined control over their data assets, HHX is especially
useful for teams collaborating on reinforcement learning models and related projects.

## Features

- **Project Management**: Create and manage projects to organize your work
- **Collection Handling**: Manage collections (buckets and tables) for storing different types of data
- **File Synchronization**: Track, stage, and push files to remote storage
- **User Authentication**: Secure user accounts with login/logout functionality
- **Storage Management**: Create and configure storage buckets with customizable settings
- **Table Management**: Create and configure database tables (this feature will be available
  soon)
- **Local Tracking**: Keep track of file changes and sync status

## Installation

```bash
# Installation instructions
# TODO: Add installation instructions
```

## Quick Start

### Initialize a Repository

```bash
# Create a new repository in the current directory
hhx init

# Create a repository linked to an existing project
hhx init --project myproject

# Create a repository with a specific collection
hhx init --project myproject --collection models
```

### Account Management

```bash
# Create a new account
hhx account create

# Log in to your account
hhx account login

# View account information
hhx account info

# View detailed account information including subscription
hhx account details
```

### Working with Projects

```bash
# Create a new project
hhx project create --name "My Project" --description "Description of my project"

# List all projects
hhx project list

# Show project details
hhx project show my-project-id

# Link repository to a project
hhx project link my-project-name
```

### Managing Collections

```bash
# List all collections
hhx collection list

# Create a new bucket collection
hhx collection create my-models --type=bucket --path=models/

# Create a new table collection with schema
hhx collection create metrics --type=table --columns="id:string:pk,timestamp:datetime,value:float"

# Show collection details
hhx collection show my-collection
```

### File Operations

```bash
# Stage files for upload
hhx stage file.txt
hhx stage directory/

# Check status of files
hhx status

# Unstage files
hhx unstage file.txt

# Push files to remote
hhx push

# Push to a specific collection
hhx push --collection=my-collection
```

### Storage Operations

```bash
# List all storage buckets
hhx storage list

# Create a new storage bucket
hhx storage create mybucket --public=false

# Show bucket details
hhx storage get mybucket
```

## Configuration

```bash
# Initialize configuration
hhx config init

# Get configuration
hhx config get

# Set configuration
hhx config set --server-url=https://api.headlesshawx.io
```

## Project Structure

HHX creates a `.hhx` directory in your repository root with the following structure:

- `.hhx/config.json` - Repository configuration
- `.hhx/index.json` - File tracking and metadata

Global configuration is stored in `~/.hhx/config.json`.

## Authentication

HHX uses token-based authentication. Tokens are stored securely in the global config directory.

## Advanced Usage

### Working with Collections

Collections in HHX come in two types:

- **Buckets**: For general file storage
- **Tables**: For structured data with schema

Linking a collection to a remote bucket:

```bash
hhx collection link my-collection --bucket=remote-bucket --create
```

### Working with Storage

Create a bucket with file size limits and content restrictions:

```bash
hhx storage create assets --file-size-limit=10485760 --allowed-mime-types="image/jpeg,image/png,application/pdf"
```

### Working with Database Tables (available soon)

Create a table with custom columns and column types:

```bash
# TODO: Implement the tables
```

## Contribution

Contributions are welcome! Please feel free to submit a Pull Request.

## License

The contents of this repository are licensed under the MIT License.