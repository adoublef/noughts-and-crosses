# Authentication Using Magic Links

## API Overview

```http
# Register a new user with email address
POST /reg/v0/users -b {email} [*]

# Delete a user from database (cascade) & email verification required
DELETE /reg/v0/users/:id [*]

# Get user profiles from database
GET /reg/v0/users/profile [*]

# Get profile for a specific user
GET /reg/v0/users/profile/:id [*]

# Set user profile information
POST /reg/v0/users/profile/:id -b {name} [*]

# Update user profile image (optional)
PUT /reg/v0/users/profile/:id/avatar [*]

# Verify email address for registration
# Make request with Auth bearer?
# CLI_URI: www.example.com/registration/verify?token={token}
GET /reg/v0/users/verify [*]

# Login with magic link
POST /auth/v0/login -b {email}

# Verify email address for login, generate session token
# Make request with Auth bearer?
# CLI_URI: www.example.com/login/verify?token={token}
GET /auth/v0/login

# Delete current session
DELETE /auth/v0/login

# Refresh session token when it expires or page refresh
GET /auth/v0/token
```

## Resources

- [magic links: all you need to know](https://www.smtp2go.com/blog/magic-links/)
- [a guide to magic links](https://workos.com/blog/a-guide-to-magic-links)
- [work os](https://workos.com/docs/magic-link/1-add-magic-link-to-your-app/add-a-callback-endpoint)
