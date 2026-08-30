# Contributing to Agent Ledger

Thank you for your interest in contributing to Agent Ledger! This document provides guidelines and instructions for contributing to the project.

## Code of Conduct

Be respectful, inclusive, and professional. We're building a welcoming community for everyone.

## Getting Started

### Prerequisites
- Go 1.21 or later
- Git
- Basic knowledge of Go and the Agent Ledger architecture

### Development Setup

```bash
# Clone the repository
git clone https://github.com/rahumanrahuu/agent-ledger.git
cd agent-ledger

# Create a feature branch
git checkout -b your-name/feature-description

# Install dependencies
go mod download

# Build the project
go build ./...

# Run tests
go test ./...
```

## Development Workflow

### 1. Find or Create an Issue

- Check [existing issues](https://github.com/rahumanrahuu/agent-ledger/issues)
- For new features, create an issue describing the problem/enhancement
- Discuss the approach with maintainers before starting work

### 2. Create a Feature Branch

```bash
git checkout -b your-name/feature-description
```

Branch naming convention:
- `feature/`: New features (`your-name/feat-feature-name`)
- `fix/`: Bug fixes (`your-name/fix-bug-name`)
- `docs/`: Documentation (`your-name/docs-topic`)
- `refactor/`: Code refactoring (`your-name/refactor-component`)

### 3. Make Your Changes

#### Code Style
- Follow Go conventions
- Use meaningful variable and function names
- Add comments for non-obvious logic
- Keep functions small and focused
- Aim for simplicity over cleverness

#### File Organization
```
agent-ledger/
├── cmd/                 # Command-line applications
├── internal/            # Internal packages
│   ├── api/            # API layer
│   ├── cache/          # Caching
│   ├── config/         # Configuration
│   ├── events/         # Event management
│   ├── memory/         # Memory/embeddings
│   ├── storage/        # File storage
│   └── ...
├── mcp/                # MCP tools
├── ui/                 # Web UI
└── tests/              # Integration tests
```

#### Adding New Packages

When adding a new package:

1. Create `internal/package-name/` directory
2. Implement core functionality in `package.go`
3. Add tests in `*_test.go` files
4. Include comprehensive package documentation
5. Update this CONTRIBUTING.md if adding new patterns

Example package structure:
```
internal/mypackage/
├── mypackage.go        # Main implementation
├── mypackage_test.go   # Tests
└── types.go            # Type definitions (if needed)
```

### 4. Testing

#### Unit Tests
```bash
# Run tests for specific package
go test ./internal/package-name -v

# Run all tests
go test ./... -v

# Run with coverage
go test ./... -cover
```

#### Test Guidelines
- Aim for 70%+ coverage on new code
- Test happy path and error cases
- Use table-driven tests for multiple scenarios
- Name tests descriptively: `TestFunctionName_Scenario`

Example:
```go
func TestCacheSetGet(t *testing.T) {
    c := NewCache(10, 1*time.Second)
    defer c.Stop()
    
    c.Set("key1", "value1")
    val, exists := c.Get("key1")
    
    if !exists {
        t.Errorf("Expected key to exist")
    }
    if val != "value1" {
        t.Errorf("Expected 'value1', got %v", val)
    }
}
```

### 5. Build and Commit

```bash
# Build to catch errors early
go build ./...

# Commit with descriptive message
git commit -m "feat: add caching layer for improved performance

- Implement TTL-based cache
- Add LRU eviction strategy
- Include comprehensive tests"
```

#### Commit Message Format
```
<type>: <subject>

<body>

<footer>
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`

### 6. Push and Create PR

```bash
# Push your branch
git push origin your-name/feature-description

# Create a pull request
gh pr create --title "feat: descriptive title" --body "
## Summary
Brief description of changes

## What changed
- Point 1
- Point 2

## How to test
- Step 1
- Step 2

## Related Issues
Fixes #123
"
```

#### PR Template

Every PR should include:
- **Summary**: 1-2 sentence overview
- **What changed**: Bullet points of modifications
- **How to test**: Steps to verify the changes
- **Related Issues**: Link to issue (Fixes #123)
- **Screenshots**: For UI changes (if applicable)

### 7. Code Review

- Respond to reviewer comments promptly
- Make requested changes and commit them
- Push updated code: `git push origin your-name/feature-description`
- Keep discussion professional and focused on code quality

## Areas for Contribution

### High Priority
- **CLI enhancements**: Better command-line interface
- **UI improvements**: Web dashboard enhancements
- **Performance optimizations**: Speed improvements
- **Documentation**: Better guides and examples

### Medium Priority
- **Test coverage**: Increase test coverage
- **Bug fixes**: Address open issues
- **Integration features**: New tool integrations
- **Error handling**: Better error messages

### Lower Priority
- **Code cleanup**: Refactoring non-critical code
- **Style improvements**: Code formatting
- **Comment improvements**: Better documentation

## Feature Ideas

### Suggested Features
1. **Webhook Support** - External tool integrations
2. **Real-time Collaboration** - Multi-agent coordination
3. **Advanced Reporting** - Custom report generation
4. **Git Integration** - Enhanced Git workflow features
5. **Machine Learning** - Predictive suggestions
6. **API Rate Limiting** - Request throttling
7. **Multi-tenancy** - Support multiple projects
8. **Audit Logging** - Track all changes
9. **Search Optimization** - Better full-text search
10. **Performance Metrics** - Detailed analytics

## Guidelines

### Do's ✅
- Write clean, readable code
- Include tests with your changes
- Update documentation as needed
- Follow existing code patterns
- Use meaningful commit messages
- Keep PRs focused and reasonably sized
- Communicate with maintainers
- Test your changes thoroughly

### Don'ts ❌
- Don't mix refactoring with feature changes
- Don't commit commented-out code
- Don't add unnecessary dependencies
- Don't make breaking changes without discussion
- Don't skip tests
- Don't ignore linting errors
- Don't create huge PRs (keep them under 500 lines if possible)
- Don't update version numbers (maintainers do this)

## Release Process

Maintainers handle releases. The process:

1. Collect merged PRs since last release
2. Update version in appropriate files
3. Create release notes
4. Tag release in Git
5. Build and publish binaries
6. Announce release

## Questions or Need Help?

- **Issues**: Use GitHub Issues for bugs and features
- **Discussions**: Start a discussion for questions
- **Slack**: Join our community Slack channel
- **Email**: Contact maintainers directly

## License

By contributing, you agree your code is licensed under the same license as the project.

## Recognition

Contributors are recognized in:
- Release notes
- CONTRIBUTORS.md file
- Project README
- Special thanks section

## Resources

- [Project Architecture](docs/architecture.md)
- [API Documentation](docs/api.md)
- [MCP Tools Guide](docs/mcp-tools.md)
- [Deployment Guide](docs/deployment.md)

## Troubleshooting

### Common Issues

**Tests failing locally but passing in CI:**
```bash
# Clear go cache
go clean -cache

# Run tests again
go test ./... -v
```

**Build errors after pulling:**
```bash
# Update dependencies
go mod tidy

# Rebuild
go build ./...
```

**Merge conflicts:**
```bash
# Update your branch with main
git fetch origin
git rebase origin/main

# Resolve conflicts, then continue
git rebase --continue
git push --force-with-lease origin your-name/feature
```

## Thank You! 🙏

Your contributions make Agent Ledger better for everyone. We appreciate:
- Bug reports
- Feature suggestions
- Code contributions
- Documentation improvements
- Community support

Happy coding! 🚀
