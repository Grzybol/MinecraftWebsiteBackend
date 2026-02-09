# Token Renewal Endpoint

## Overview

The token renewal endpoint allows users to refresh their JWT token without having to re-authenticate with credentials. This is useful for maintaining user sessions and avoiding frequent login prompts.

## Endpoint

```
POST /api/renew
```

## Authentication

**Required**: Valid JWT token in Authorization header

```
Authorization: Bearer <your-jwt-token>
```

## Rate Limiting

**Rate Limit**: 5 requests per millisecond

## Request

No request body required. The endpoint uses the token from the Authorization header.

## Example Request

```bash
curl -X POST "https://bestservers.fun:8443/api/renew" \
  -H "Authorization: Bearer your-jwt-token"
```

## Response

### Success Response (200 OK)

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "Token renewed successfully"
}
```

### Error Responses

#### 401 Unauthorized
```json
{
  "error": "No token provided"
}
```
or
```json
{
  "error": "token is not valid"
}
```
or
```json
{
  "error": "token has expired"
}
```
or
```json
{
  "error": "invalid token claims"
}
```

#### 429 Too Many Requests
```json
{
  "error": "Zbyt wiele zapytań do /api/renew – spróbuj ponownie za 0.0 sekund."
}
```

## Technical Details

### Token Validation

The endpoint validates the existing token before renewal:

1. **Token Format**: Validates JWT format and signature
2. **Token Expiration**: Checks if token hasn't expired (with 1-minute margin)
3. **Token Claims**: Validates required claims (user_id, exp)
4. **Token Revocation**: Checks if token isn't in the revocation list

### New Token Generation

If validation passes, a new token is generated with:

- **Same user_id**: Preserves user identity
- **New expiration**: 10 minutes from current time
- **Same algorithm**: HS256 signing
- **Same secret**: Uses application JWT secret

### Security Features

- **Rate limiting**: Prevents abuse
- **Token validation**: Ensures only valid tokens can be renewed
- **Expiration check**: Prevents renewal of expired tokens
- **Revocation check**: Prevents renewal of revoked tokens
- **Logging**: All renewal attempts are logged

## Usage Examples

### JavaScript/Fetch

```javascript
async function renewToken(currentToken) {
  const response = await fetch('/api/renew', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer ' + currentToken
    }
  });

  if (response.ok) {
    const result = await response.json();
    console.log('Token renewed:', result.token);
    return result.token;
  } else {
    const error = await response.json();
    console.error('Renewal failed:', error.error);
    throw new Error(error.error);
  }
}

// Usage
try {
  const newToken = await renewToken(currentToken);
  // Update stored token
  localStorage.setItem('token', newToken);
} catch (error) {
  // Handle error - redirect to login
  window.location.href = '/login';
}
```

### Python/Requests

```python
import requests

def renew_token(current_token):
    response = requests.post(
        'https://bestservers.fun:8443/api/renew',
        headers={'Authorization': f'Bearer {current_token}'}
    )
    
    if response.status_code == 200:
        result = response.json()
        print(f'Token renewed: {result["token"]}')
        return result['token']
    else:
        error = response.json()
        print(f'Renewal failed: {error["error"]}')
        raise Exception(error['error'])

# Usage
try:
    new_token = renew_token(current_token)
    // Update stored token
    with open('token.txt', 'w') as f:
        f.write(new_token)
except Exception as e:
    // Handle error - redirect to login
    print(f"Need to login again: {e}")
```

### Automatic Token Renewal

```javascript
// Example of automatic token renewal before API calls
async function apiCallWithTokenRenewal(endpoint, options = {}) {
  let token = localStorage.getItem('token');
  
  // Try to renew token before making API call
  try {
    const newToken = await renewToken(token);
    token = newToken;
    localStorage.setItem('token', newToken);
  } catch (error) {
    // Token renewal failed, redirect to login
    window.location.href = '/login';
    return;
  }
  
  // Make API call with renewed token
  const response = await fetch(endpoint, {
    ...options,
    headers: {
      ...options.headers,
      'Authorization': 'Bearer ' + token
    }
  });
  
  return response;
}
```

## Best Practices

### When to Renew

- **Proactive renewal**: Renew tokens before they expire (e.g., 1-2 minutes before)
- **On 401 errors**: If an API call returns 401, try renewing the token
- **Periodic renewal**: Set up a timer to renew tokens every 8-9 minutes

### Error Handling

- **Expired tokens**: Redirect to login page
- **Invalid tokens**: Clear stored token and redirect to login
- **Rate limiting**: Implement exponential backoff for retries

### Security Considerations

- **Token storage**: Store tokens securely (httpOnly cookies, secure localStorage)
- **Token transmission**: Always use HTTPS
- **Token cleanup**: Clear tokens on logout
- **Token validation**: Validate tokens on both client and server side

## Notes

- Tokens have a 10-minute lifetime
- Renewal is only possible for tokens that haven't expired
- The endpoint requires a valid, non-expired, non-revoked token
- All renewal attempts are logged for security monitoring
- Rate limiting prevents abuse of the renewal endpoint 