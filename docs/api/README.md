# API Documentation

## Overview

This API provides endpoints for managing player data, equipment, and server interactions for the BestServers.fun platform.

## Setup

Before using the API, you need to configure the application. See [Setup Instructions](../SETUP.md) for detailed configuration steps.

## Authentication

All protected endpoints require a JWT token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

## Rate Limiting

Most endpoints are protected by rate limiting to prevent abuse. Rate limits vary by endpoint:

- Login: 1 request per millisecond
- Equipment endpoints: 5 requests per millisecond
- Payment endpoints: 10-10000 requests per millisecond (varies by type)
- General endpoints: 10 requests per millisecond

## Base URL

```
https://bestservers.fun:8443
```

## CORS

The API supports CORS for the following domains:
- https://bestservers.fun
- https://boxpvp.top

## Error Responses

All endpoints return errors in the following format:

```json
{
  "error": "Error description"
}
```

Common HTTP status codes:
- 200: Success
- 400: Bad Request
- 401: Unauthorized
- 403: Forbidden
- 404: Not Found
- 429: Too Many Requests
- 500: Internal Server Error

## Available Endpoints

### Authentication
- `POST /api/login` - User login
- `GET /api/user` - Get user information
- `POST /api/logout` - User logout
- `POST /api/renew` - Renew JWT token

### Equipment
- `GET /api/getBackpack` - Get player backpack
- `GET /api/getArmor` - Get player armor
- `GET /api/getHotbar` - Get player hotbar
- `GET /api/getEcoBalance` - Get player eco balance

### Payments
- `GET /api/checkPlayerBalance` - Check player website balance
- `POST /api/processCodePayment` - Process payment with code
- `POST /api/processPointsPayment` - Process payment with points

### Game Data
- `GET /api/usergroup` - Get user group information
- `POST /api/checkPlayerOnline` - Check if player is online
- `GET /api/getEloRanking` - Get ELO ranking
- `GET /api/getPlayerElo` - Get player ELO
- `GET /api/getPlayerByRank` - Get player by rank
- `GET /stats/:world` - Get aggregated world statistics (public). See [World Stats](world-stats.md).

### Mobile App
- `GET /api/barcodeinfo` - Get barcode information

### Clicker Game
- `POST /api/progress/sync` - Sync player progress
- `GET /api/progress` - Get player progress
- `GET /api/leaderboard` - Get player leaderboard

## Configuration

Server tokens and other sensitive configuration are stored in `config.go`. This file is excluded from version control for security reasons. See [Setup Instructions](../SETUP.md) for configuration details.