# Trakt.tv API Integration Guide

This document explains how Export Trakt for Letterboxd interacts with the Trakt.tv API, providing details on authentication, endpoints used, rate limiting, and best practices.

## Table of Contents

- [API Overview](#api-overview)
- [Authentication](#authentication)
- [Endpoints Used](#endpoints-used)
- [Rate Limiting](#rate-limiting)
- [Error Handling](#error-handling)
- [Improving API Performance](#improving-api-performance)
- [API Client Implementation](#api-client-implementation)
- [Extending API Functionality](#extending-api-functionality)

## API Overview

The Export Trakt for Letterboxd application interacts with the [Trakt.tv API](https://trakt.docs.apiary.io/) to retrieve user watch history, ratings, and other data. Trakt.tv provides a RESTful API that allows access to various resources related to movies, TV shows, and user activities.

Key features of the Trakt.tv API integration:

- OAuth 2.0 authentication for secure API access
- Retrieval of watched movies with dates and ratings
- Access to watchlist and collection data
- Pagination support for handling large datasets
- Rate limiting to comply with Trakt.tv API policies

## Authentication

The application uses OAuth 2.0 for authentication with the Trakt.tv API.

### Authentication Flow

1. **Register Trakt.tv Application**

   - Go to [https://trakt.tv/oauth/applications](https://trakt.tv/oauth/applications)
   - Register a new application to get client ID and client secret
   - Set redirect URI to `urn:ietf:wg:oauth:2.0:oob` for command-line applications

2. **Authorization Code Grant Flow**

   - The application directs the user to a Trakt.tv authorization URL
   - User authorizes the application on Trakt.tv
   - User receives an authorization code
   - Application exchanges the code for an access token

3. **Token Management**
   - Access tokens are valid for 3 months
   - Refresh tokens are valid for indefinite use
   - The application stores tokens securely in the token file
   - Automatic token refresh when the access token expires

### Implementation Details

The authentication process is implemented in `pkg/api/auth.go`. Key functions include:

- `GetAuthorizationURL`: Generates the URL for the user to authorize the application
- `GetAccessToken`: Exchanges the authorization code for an access token
- `RefreshAccessToken`: Refreshes an expired token
- `LoadTokenFromFile`: Loads saved token data
- `SaveTokenToFile`: Securely saves token data

## Endpoints Used

The application uses the following Trakt.tv API endpoints:

### Movies Endpoints

- `GET /sync/watched/movies`: Retrieve watched movies history
- `GET /sync/ratings/movies`: Retrieve movie ratings
- `GET /sync/watchlist/movies`: Retrieve movie watchlist
- `GET /sync/collection/movies`: Retrieve movie collection (optional)

### User Endpoints

- `GET /users/{username}/watched/movies`: Alternative method to get watched movies
- `GET /users/{username}/ratings/movies`: Alternative method to get movie ratings

### Search Endpoints

- `GET /search/movie`: Search for movies (used for verification and enrichment)

### Additional Endpoints

- `GET /movies/{id}`: Get detailed information about a specific movie
- `GET /movies/{id}/aliases`: Get alternative titles for a movie

## Rate Limiting

The Trakt.tv API has rate limits that must be respected:

- 1,000 requests per 5 minutes (average of 3.33 requests per second)
- 5 requests per second for personal websites
- 50 requests per second for Trakt.tv clients

### How Rate Limiting is Implemented

The application implements rate limiting in `pkg/api/client.go` through:

1. **Client-side Throttling**

   - Limiting request frequency to stay below limits
   - Configurable rate limit in the application settings

2. **Response Header Monitoring**

   - Monitoring `X-Ratelimit-Remaining` and `X-Ratelimit-Reset` headers
   - Adjusting request timing based on remaining limits

3. **Exponential Backoff**
   - Implementing backoff strategy for rate limit errors
   - Starting with a short delay and progressively increasing

## Error Handling

The API client implements robust error handling for various error scenarios:

### Error Types

- **Authentication Errors**: Issues with API keys or tokens
- **Rate Limiting Errors**: When API limits are exceeded
- **Network Errors**: Connectivity issues to the Trakt.tv API
- **Server Errors**: 5xx responses from the Trakt.tv API
- **Client Errors**: Invalid requests (4xx responses)
- **Parsing Errors**: Issues with parsing API responses

### Retry Strategy

The application implements an intelligent retry strategy:

- Automatic retry for transient errors (5xx, network issues)
- Exponential backoff with jitter for rate limit errors
- Configurable maximum retry attempts
- Different strategies based on error type

## Improving API Performance

The application optimizes API usage through several strategies:

### Caching

- In-memory caching of frequently accessed data
- Disk caching of API responses with configurable TTL
- Separate cache for metadata that changes infrequently

### Pagination Optimization

- Proper handling of paginated responses
- Batch processing of pages to minimize API calls
- Dynamic page size adjustment based on endpoint

### Parallel Requests

- Concurrent requests for independent data
- Configurable concurrency limits
- Coordination to prevent rate limit issues

## API Client Implementation

The Trakt.tv API client is implemented in the `pkg/api` package. The main components include:

### TraktClient Interface

```go
type TraktClient interface {
    GetWatchedMovies() ([]Movie, error)
    GetRatedMovies() ([]RatedMovie, error)
    GetWatchlistMovies() ([]Movie, error)
    GetCollectionMovies() ([]Movie, error)
    SearchMovie(query string) ([]SearchResult, error)
    GetMovieDetails(id string) (*MovieDetails, error)
}
```

### Client Configuration

```go
type Config struct {
    ClientID     string
    ClientSecret string
    RedirectURI  string
    TokenFile    string
    Timeout      time.Duration
    MaxRetries   int
    RateLimit    int
}
```

### Data Structures

The client uses various data structures to represent Trakt.tv entities:

- `Movie`: Basic movie information
- `RatedMovie`: Movie with user rating
- `WatchedMovie`: Movie with watch date
- `SearchResult`: Search result item
- `MovieDetails`: Detailed movie information

## Extending API Functionality

To extend the API functionality, follow these steps:

### Adding New Endpoints

1. Add method to the `TraktClient` interface
2. Implement the method in `client.go`
3. Add appropriate data structures
4. Add unit tests for the new functionality

### Example: Adding TV Shows Support

```go
// Add to TraktClient interface
GetWatchedShows() ([]Show, error)

// Implement in client.go
func (c *client) GetWatchedShows() ([]Show, error) {
    var shows []Show
    url := fmt.Sprintf("%s/sync/watched/shows", baseURL)

    err := c.makeRequest("GET", url, nil, &shows)
    if err != nil {
        return nil, fmt.Errorf("failed to get watched shows: %w", err)
    }

    return shows, nil
}
```

## Best Practices

When working with the Trakt.tv API:

1. **Respect Rate Limits**: Always implement proper rate limiting
2. **Handle Pagination**: Most endpoints return paginated results
3. **Implement Caching**: Minimize API calls with appropriate caching
4. **Secure Tokens**: Store access and refresh tokens securely
5. **Implement Proper Error Handling**: Log and handle all errors appropriately
6. **Keep Authentication Up-to-Date**: Implement token refresh logic
7. **Use Extended Info**: Use `?extended=full` parameter when needed
8. **Verify User Permissions**: Check that the user has authorized necessary scopes

## Troubleshooting

Common issues and solutions:

### Authentication Failed

**Problem**: Unable to authenticate with the Trakt.tv API.
**Solution**:

- Verify that your client ID and client secret are correct
- Ensure the user has granted permission to your application
- Check if your token file is corrupted or has invalid permissions

### Rate Limiting

**Problem**: Hitting rate limits consistently.
**Solution**:

- Increase the time between requests in your configuration
- Implement more aggressive caching
- Consider batch processing data less frequently

### Missing Data

**Problem**: Some expected data is missing from the API responses.
**Solution**:

- Ensure the user's Trakt.tv profile is public
- Check if the user has the expected data in their Trakt.tv account
- Verify you're using the correct API endpoints

## References

- [Trakt.tv API Documentation](https://trakt.docs.apiary.io/)
- [OAuth 2.0 Specification](https://oauth.net/2/)
- [HTTP Status Codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
