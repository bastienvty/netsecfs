# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

NetSecFS is a shared encrypted network file system built in Go that provides secure file sharing capabilities with user authentication and directory sharing features. The system uses FUSE (Filesystem in Userspace) to mount an encrypted filesystem and SQLite databases for metadata and object storage.

## Build Commands

```bash
# Build the application
go build .

# Clean dependencies 
go mod tidy

# Run the built binary (requires initialization first)
./netsecfs --meta meta.db /tmp/nsfs
```

## Initialization and Usage

```bash
# Initialize filesystem with storage and metadata databases
./netsecfs init --storage data.db --meta meta.db myfs

# Mount the filesystem with interactive CLI
./netsecfs --meta meta.db /tmp/nsfs

# Inside the CLI, available commands:
# signup <username> <password> - Create new user
# login <username> <password> - Login existing user  
# mount - Mount filesystem for current user
# umount - Unmount filesystem
# share <folder_path> <user> - Share directory with another user
# unshare <folder_path> <user> - Remove directory sharing
# passwd <new_password> - Change user password
# logout - Logout current user
# help - Show available commands
# exit - Exit application
```

## Architecture

### Core Components

- **CLI Layer** (`internal/cli/`): Interactive command-line interface with user management and filesystem operations
- **Database Layer** (`internal/db/`): 
  - `meta/`: User management, filesystem metadata, sharing permissions using SQLite/XORM
  - `object/`: Encrypted file data storage using LevelDB
- **Filesystem Layer** (`internal/fs/`): FUSE filesystem implementation with encryption/decryption
- **Crypto Layer** (`internal/crypto/`): Encryption utilities for file content and user data
- **Commands** (`cmd/`): Cobra CLI framework setup with init and mount commands

### Key Data Flow

1. Files are encrypted using per-file keys before storage
2. File keys are encrypted with user's root key derived from password
3. Metadata (filenames, permissions, sharing) stored in SQLite database
4. Encrypted file content stored in LevelDB object storage
5. Directory sharing uses public key cryptography for key exchange

### Database Schema

- **Users**: Authentication with salted password hashes, public/private key pairs
- **Nodes**: File/directory metadata with encryption keys
- **Edges**: Parent-child relationships in directory structure  
- **Shared**: Directory sharing permissions with encrypted keys per user

## Platform Requirements

- Go 1.22.2 or later
- Ubuntu: `fuse build-essential libsqlite3-dev` packages
- macOS: macFUSE installation required
- Supports both ARM and x86 architectures

## Important Notes

- This is a proof of concept - not for production use
- Mount operations require proper FUSE setup on the system
- Uses SQLite for metadata and LevelDB for object storage
- All file content is encrypted before storage
- Directory sharing requires both users to exist in the system