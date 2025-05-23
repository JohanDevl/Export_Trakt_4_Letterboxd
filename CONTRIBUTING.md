# Contributing to Export Trakt 4 Letterboxd

Thank you for your interest in contributing to Export Trakt 4 Letterboxd! This document provides guidelines and instructions for contributing to this project.

## Code of Conduct

By participating in this project, you agree to abide by our code of conduct. Please be respectful and considerate of others.

## How to Contribute

### Reporting Bugs

If you find a bug, please create an issue using the bug report template. Be sure to include:

- A clear description of the bug
- Steps to reproduce the issue
- Expected behavior
- Screenshots if applicable
- Your environment details

### Suggesting Enhancements

If you have an idea for an enhancement, please create an issue using the feature request template. Be sure to include:

- A clear description of the feature
- The motivation for the feature
- Any alternative solutions you've considered

### Pull Requests

1. Fork the repository
2. Create a new branch for your feature or bugfix
3. Make your changes
4. Test your changes
5. Submit a pull request

Please follow these guidelines for your pull requests:

- Follow the coding style of the project
- Write clear commit messages
- Include tests for your changes
- Update documentation as needed
- Reference any related issues

## Development Setup

### Prerequisites

- A Trakt.tv account
- A Trakt.tv application (Client ID and Client Secret)
- `jq` and `curl` installed on your system

### Local Setup

1. Clone the repository:

   ```bash
   git clone https://github.com/YOUR-USERNAME/Export_Trakt_4_Letterboxd.git
   cd Export_Trakt_4_Letterboxd
   ```

2. Make the scripts executable:

   ```bash
   chmod +x Export_Trakt_4_Letterboxd.sh setup_trakt.sh
   ```

3. Configure Trakt authentication:
   ```bash
   ./setup_trakt.sh
   ```

### Docker Setup

1. Clone the repository:

   ```bash
   git clone https://github.com/YOUR-USERNAME/Export_Trakt_4_Letterboxd.git
   cd Export_Trakt_4_Letterboxd
   ```

2. Start the container:

   ```bash
   docker compose up -d
   ```

3. Configure Trakt authentication:
   ```bash
   docker compose exec trakt-export ./setup_trakt.sh
   ```

## Testing

Before submitting a pull request, please test your changes thoroughly. This includes:

- Testing the main functionality
- Testing edge cases
- Ensuring Docker compatibility if applicable

## Documentation

If you're changing functionality, please update the relevant documentation in the `wiki/` directory.

## License

By contributing to this project, you agree that your contributions will be licensed under the same license as the project.

Thank you for contributing to Export Trakt 4 Letterboxd!
