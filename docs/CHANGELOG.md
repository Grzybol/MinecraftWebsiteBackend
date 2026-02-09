# Changelog

All notable changes to the BestServers.fun API will be documented in this file.

## [Unreleased]

### Added
- New endpoint `/api/getEcoBalance` for retrieving player economic balance from Minecraft servers
- New endpoint `/api/renew` for JWT token renewal without re-authentication
- New clicker game endpoints for player progress management:
  - `POST /api/progress/sync` - Sync player progress
  - `GET /api/progress` - Get player progress  
  - `GET /api/leaderboard` - Get player leaderboard
- Comprehensive API documentation in `/docs/api/` directory
- Detailed documentation for eco balance endpoint (`/docs/api/eco-balance.md`)
- Detailed documentation for token renewal endpoint (`/docs/api/renew.md`)
- Detailed documentation for clicker game endpoints (`/docs/api/clicker.md`)
- Equipment endpoints documentation (`/docs/api/equipment.md`)
- General API overview documentation (`/docs/api/README.md`)
- Setup instructions (`/docs/SETUP.md`)

### Security & Configuration
- Moved sensitive configuration to `config.go` (excluded from version control)
- Added `.gitignore` to exclude sensitive files
- Removed sensitive data (server tokens, database credentials) from documentation
- Centralized all configuration in `config.go` including ServerTokens
- Added setup instructions for secure configuration
- **NEW**: Implemented comprehensive anti-cheat verification system for clicker game
  - Maximum 25 clicks per second limit
  - Resource gain validation based on player level and upgrades
  - Rollback detection and time validation
  - 5% safety margin for network delays
  - Detailed verification failure reporting

### Technical Details
- Eco balance endpoint communicates with `/tmp/plugin_minecraft_ecobalance.sock`
- Token renewal endpoint validates existing tokens and generates new ones with 10-minute lifetime
- Clicker game endpoints manage player progress in MongoDB `player_progress` collection
- **NEW**: Anti-cheat verification system with mathematical validation of resource gains
- Uses same authentication and rate limiting patterns as existing equipment endpoints
- Rate limit: 5-10 requests per millisecond (varies by endpoint)
- Requires JWT authentication
- Supports all existing servers (boxpvp, survival, skygen)
- Comprehensive logging and error handling

### Files Added
- `ecoBalanceHandler.go` - New handler for eco balance endpoint
- `clickerHandler.go` - New handlers for clicker game endpoints
- `docs/api/README.md` - General API documentation
- `docs/api/eco-balance.md` - Eco balance endpoint documentation
- `docs/api/renew.md` - Token renewal endpoint documentation
- `docs/api/clicker.md` - Clicker game endpoints documentation
- `docs/api/equipment.md` - Equipment endpoints documentation
- `docs/SETUP.md` - Setup and configuration instructions
- `docs/CHANGELOG.md` - This changelog file
- `.gitignore` - Git ignore rules for sensitive files

### Files Modified
- `main.go` - Added new routes for `/api/getEcoBalance`, `/api/renew`, and clicker game endpoints
- `config.go` - Centralized all configuration including ServerTokens
- `utils.go` - Removed ServerTokens (moved to config.go)
- `auth.go` - Added RenewJWT function for token renewal
- `handlers.go` - Added RenewTokenHandler for token renewal endpoint
- `models.go` - Added PlayerProgress, Upgrade, and verification structures for clicker game
- `clickerHandler.go` - Added comprehensive anti-cheat verification system
- Documentation files - Removed sensitive data and added new endpoints with verification details 