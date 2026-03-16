# Access Control Module

## Overview
The Access Control module provides a comprehensive system for managing permissions and access rights to resources in AEShield. It implements the access control specification with support for public, private, and whitelist access modes.

## Architecture

### Models
- `AccessRule`: Represents an access control rule with resource ID, type, owner, access mode, and whitelist
- `AccessMode`: Enum for access modes (Public, Private, Whitelist)
- `CheckAccessRequest`: Structure for access verification requests
- `CheckAccessResult`: Structure for access verification results

### Repository Layer
- `AccessControlRepository`: Interface defining data access operations
- `MongoAccessControlRepository`: MongoDB implementation of the repository interface
- Supports CRUD operations for access rules
- Provides ownership verification and whitelist management

### Service Layer
- `AccessControlService`: Business logic layer implementing access control rules
- Manages access rule lifecycle (create, update, delete)
- Implements access verification logic based on access mode
- Provides whitelist management operations

### Handler Layer
- `Handler`: HTTP handler implementing API endpoints
- Provides RESTful endpoints for access control operations
- Implements authentication and authorization checks

## Features

### Access Modes
1. **Public**: Anyone with the link can access the resource
2. **Private**: Only the owner can access the resource
3. **Whitelist**: Only specified users/emails can access the resource

### Key Operations
- Create access rules for resources
- Update access modes and whitelists
- Check access permissions for users
- Add/remove users from whitelists
- Verify resource ownership

## API Endpoints

- `POST /access/rules` - Create new access rule
- `GET /access/rules/{resource_id}` - Get access rule for resource
- `PATCH /access/rules/{resource_id}` - Update access rule
- `DELETE /access/rules/{resource_id}` - Delete access rule
- `POST /access/check` - Check access permissions
- `POST /access/whitelist` - Add to whitelist
- `DELETE /access/whitelist` - Remove from whitelist

## Security
- Implements proper authentication checks using JWT tokens
- Enforces ownership validation for privileged operations
- Provides secure access verification mechanisms
- Handles sensitive whitelist data appropriately

## Integration
The module is designed to work seamlessly with the existing AEShield architecture and can be integrated with file management, user management, and other resource management systems.