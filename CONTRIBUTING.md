# Contributing to the Standalone Google Drive Client

Thank you for considering contributing to this project! Here's how you can help.

## Code of Conduct

Please be respectful and considerate of others when contributing to this project.

## How Can I Contribute?

### Reporting Bugs

- Check if the bug has already been reported by searching the issues
- If not, create a new issue using the bug report template
- Include as much relevant information as possible:
  - Steps to reproduce
  - Expected behavior
  - Actual behavior
  - Error messages and stack traces
  - Environment details

### Suggesting Features

- Check if the feature has already been suggested
- If not, create a new issue using the feature request template
- Explain why this feature would be useful

### Pull Requests

1. Fork the repository
2. Create a new branch for your feature or bug fix
3. Make your changes
4. Add or update tests as needed
5. Run the test suite to ensure everything passes
6. Push your branch to your fork
7. Submit a pull request to the main repository

## Development Setup

1. Clone the repository
2. Install Go (version 1.18 or later recommended)
3. Run `go get` to install dependencies
4. Set up testing credentials (see docs/oauth_guide.md)

## Coding Standards

- Follow Go's standard formatting (run `go fmt` before committing)
- Ensure all tests pass (run `go test ./...`)
- Document new functions and types
- Add tests for new functionality

## Testing

Please ensure all tests pass before submitting a pull request:

```bash
go test ./...
```

For more detailed testing information, see the testing documentation.

## Documentation

If you're changing functionality or adding features, please update the documentation accordingly:

- README.md for user-facing changes
- Code comments for API documentation
- docs/ directory for detailed guides

Thank you for contributing!
