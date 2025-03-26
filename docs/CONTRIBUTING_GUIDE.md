# Contributing Guide

This document provides detailed information for contributors to the Export Trakt for Letterboxd project. It expands on the information in the main [CONTRIBUTING.md](../CONTRIBUTING.md) file with more specific guidelines.

## Table of Contents

- [Development Workflow](#development-workflow)
- [Code Style and Standards](#code-style-and-standards)
- [Testing Guidelines](#testing-guidelines)
- [Documentation Guidelines](#documentation-guidelines)
- [Internationalization (i18n)](#internationalization-i18n)
- [Pull Request Process](#pull-request-process)
- [Issue Tracking](#issue-tracking)

## Development Workflow

We follow a GitHub Flow-based development workflow:

1. **Fork and Clone**: Fork the repository and clone it locally
2. **Branch**: Create a feature branch from `main`
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **Develop**: Make your changes following our guidelines
4. **Test**: Ensure all tests pass
   ```bash
   go test -v ./...
   ```
5. **Commit**: Use clear, descriptive commit messages
   ```bash
   git commit -m "Add feature: description of the change"
   ```
6. **Push**: Push your branch to your fork
   ```bash
   git push origin feature/your-feature-name
   ```
7. **Pull Request**: Open a PR against the `main` branch

## Code Style and Standards

### Go Code Style

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` or `goimports` to format your code before committing
- Run `golint` and `go vet` to catch common issues
- Maintain package-level documentation for all exported types and functions
- Follow idiomatic Go practices:
  - Use meaningful variable names
  - Return errors rather than using panics
  - Minimize use of global variables
  - Use interfaces for dependency injection

### Naming Conventions

- **Files**: Use snake_case for filenames
- **Packages**: Use lowercase, single-word names
- **Functions/Methods**: Use CamelCase for exported functions, camelCase for internal functions
- **Variables**: Use camelCase for variables
- **Constants**: Use UPPER_SNAKE_CASE for constants

## Testing Guidelines

### Unit Tests

- Write unit tests for all new functionality
- Aim for at least 80% test coverage for new code
- Use table-driven tests where appropriate
- Keep tests focused and fast
- Use mocks for external dependencies

Example unit test:

```go
func TestGetWatchedMovies(t *testing.T) {
    // Setup test data
    mockClient := mocks.NewMockTraktClient()

    // Run the function under test
    movies, err := GetWatchedMovies(mockClient)

    // Assert expectations
    assert.NoError(t, err)
    assert.Len(t, movies, 2)
    assert.Equal(t, "Movie Title", movies[0].Title)
}
```

### Integration Tests

- Integration tests should verify end-to-end functionality
- Use the `tests/integration` directory for integration tests
- Avoid external dependencies in integration tests where possible

## Documentation Guidelines

### Code Documentation

- Document all exported functions, types, and constants
- Follow godoc conventions for documentation comments
- Include examples for complex functions

Example:

```go
// ExportMovies exports the user's watched movies to a CSV file compatible with Letterboxd.
// It retrieves the data from Trakt.tv and formats it according to Letterboxd's import format.
//
// Parameters:
//   - client: A configured Trakt client for API access
//   - outputPath: The path where the CSV file should be saved
//
// Returns:
//   - The number of movies exported
//   - Any error encountered during the export process
func ExportMovies(client api.TraktClient, outputPath string) (int, error) {
    // Implementation
}
```

### Project Documentation

- Update README.md when adding significant features
- Keep documentation in the `docs/` directory up-to-date
- Use Markdown for all documentation files
- Include screenshots or diagrams when they help explain concepts

## Internationalization (i18n)

When adding new user-facing strings:

1. Add the new string to the `locales/en.json` file under the appropriate section
2. Follow the existing format with appropriate message IDs
3. Use meaningful message IDs that describe the purpose of the message
4. Include any placeholders with descriptive names

Example:

```json
{
  "export": {
    "movies_exported": "Successfully exported {{count}} movies to {{file}}",
    "no_movies_found": "No movies found in your Trakt.tv history"
  }
}
```

## Pull Request Process

1. Ensure all tests pass locally before submitting
2. Update documentation for any changed functionality
3. Add an entry to the CHANGELOG.md in the "Unreleased" section
4. Link to any related issues using GitHub's keyword syntax (`Fixes #123`)
5. Request review from at least one maintainer
6. Be responsive to feedback and make requested changes

## Issue Tracking

When creating issues:

- Use the appropriate issue template
- Provide as much detail as possible
- For bugs, include steps to reproduce, expected behavior, and actual behavior
- For features, explain the motivation and potential implementation
- Label issues appropriately

## Code Review

When reviewing others' contributions:

- Be respectful and constructive
- Focus on the code, not the person
- Explain your reasoning for requested changes
- Approve PRs when they meet project standards, even if you would implement differently

## License Compliance

- Ensure all new files include the project's license header
- Do not introduce dependencies with incompatible licenses
- Document third-party code usage

Thank you for contributing to Export Trakt for Letterboxd!
