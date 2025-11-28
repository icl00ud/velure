# SonarCloud Setup - Product Service

This document describes the SonarCloud integration for code quality and test coverage analysis.

## Overview

The Product Service is configured to automatically send test coverage reports and code quality metrics to SonarCloud on each CI/CD pipeline run.

## Configuration Files

### 1. `sonar-project.properties`

Located in the root of the product-service directory, this file contains the SonarCloud project configuration:

- **Project Key**: `velure-product-service`
- **Organization**: `icl00ud`
- **Coverage Report Path**: `coverage.out`
- **Source Exclusions**: Test files, vendor directory, generated code

### 2. GitHub Actions Workflow

The `.github/workflows/go-service.yml` workflow includes:

1. **Test Execution**: Runs all tests with coverage reporting
   ```bash
   go test ./... -coverprofile=coverage.out -covermode=atomic -v
   ```

2. **SonarCloud Analysis**: Automatically uploads code and coverage data
   - Uses the official `SonarSource/sonarcloud-github-action`
   - Requires `SONAR_TOKEN` secret to be configured

3. **Coverage Artifact**: Saves coverage report for 30 days

## Setup Instructions

### Prerequisites

1. **SonarCloud Account**: Create an account at https://sonarcloud.io
2. **Organization**: Create or join an organization (e.g., `velure`)
3. **Project**: Create a new project with key `velure-product-service`

### GitHub Repository Configuration

1. **Add SonarCloud Secret**:
   - Go to your GitHub repository settings
   - Navigate to `Settings > Secrets and variables > Actions`
   - Add a new repository secret:
     - Name: `SONAR_TOKEN`
     - Value: Your SonarCloud token (get it from SonarCloud.io > My Account > Security)

2. **Verify Workflow Configuration**:
   - Ensure `.github/workflows/ci-cd.yml` passes the `SONAR_TOKEN` secret to the Go service workflow
   - The SonarCloud step uses `continue-on-error: true` so builds succeed even without the token
   - If token is not configured, SonarCloud step will be skipped silently

### SonarCloud Project Configuration

1. **Import Repository**:
   - In SonarCloud, import your GitHub repository
   - Select the organization and project key to match `sonar-project.properties`

2. **Configure Quality Gate** (Optional):
   - Set minimum coverage threshold (current: 75% for testable code)
   - Configure code smell and bug thresholds

## Coverage Reporting

### Current Coverage (as of latest commit)

```
Total Internal Package Coverage: 59.7%

Package Breakdown:
- Services Layer:    100.0% ✅
- Handlers Layer:    94.0%  ✅
- Middleware:        90.9%  ✅
- Config:            86.2%  ✅
- Models:            N/A (data structures)
- Metrics:           N/A (definitions)
- Repository:        0.0% (requires integration tests)
```

### Understanding the Coverage

**High Coverage Components** (Business Logic):
- All critical business logic has >= 75% coverage
- Services layer has 100% coverage (all business rules tested)
- Handlers layer has 94% coverage (comprehensive endpoint testing)

**Repository Layer** (0% coverage):
- Requires MongoDB and Redis instances for proper testing
- These are integration tests, not unit tests
- Would be implemented using testcontainers in production
- The repository layer is data access only (no business logic)

## Viewing Results

### In Pull Requests

SonarCloud will automatically:
- Comment on PRs with quality gate status
- Show coverage changes (increase/decrease)
- Highlight new code smells, bugs, or vulnerabilities

### In SonarCloud Dashboard

Access detailed metrics at:
```
https://sonarcloud.io/project/overview?id=velure-product-service
```

View:
- Code coverage trends
- Code quality metrics
- Technical debt
- Security vulnerabilities
- Code duplication

## Local Testing

To generate coverage report locally:

```bash
cd services/product-service

# Run tests with coverage
go test ./... -coverprofile=coverage.out -covermode=atomic

# View coverage summary
go tool cover -func=coverage.out

# View coverage in browser
go tool cover -html=coverage.out
```

## Troubleshooting

### Coverage Not Appearing in SonarCloud

1. Check that `coverage.out` is being generated in the workflow
2. Verify the path in `sonar-project.properties` matches the generated file
3. Ensure `SONAR_TOKEN` secret is configured correctly
4. Check workflow logs for SonarCloud scanner errors
5. Note: The SonarCloud step uses `continue-on-error: true`, so check the step output even if the workflow succeeds

### Quality Gate Failing

1. Review the specific metrics that are failing
2. Check if coverage dropped below threshold
3. Review new code smells or bugs introduced
4. Fix issues and push new commit

### Workflow Not Running

1. Verify the service path filter in `.github/workflows/ci-cd.yml`
2. Check that changes are in the `services/product-service/**` directory
3. Ensure workflow has proper permissions

## Best Practices

1. **Maintain Coverage**: Keep critical business logic coverage >= 75%
2. **Review Quality Reports**: Check SonarCloud feedback on each PR
3. **Fix Issues Early**: Address code smells and bugs before merging
4. **Document Skipped Tests**: Clearly mark integration tests that require external dependencies
5. **Monitor Trends**: Track coverage and quality metrics over time

## References

- [SonarCloud Documentation](https://docs.sonarcloud.io/)
- [Go Coverage Guide](https://go.dev/blog/cover)
- [GitHub Actions with SonarCloud](https://github.com/SonarSource/sonarcloud-github-action)
