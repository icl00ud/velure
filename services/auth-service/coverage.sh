#!/bin/bash

# Script to generate test coverage excluding generated code (mocks and testutil)
# This gives a more accurate representation of actual code coverage

echo "🧪 Running tests with coverage..."
go test ./... -coverprofile=coverage_full.out -covermode=atomic

if [ $? -ne 0 ]; then
    echo "❌ Tests failed"
    exit 1
fi

echo "🔍 Filtering out generated code (mocks and testutil)..."
grep -v "/internal/mocks/" coverage_full.out | grep -v "/internal/testutil/" > coverage.out

echo ""
echo "📊 Coverage Report (excluding generated code):"
echo "================================================"
go tool cover -func=coverage.out | grep -E "^velure-auth-service/" | grep -v "100.0%"
echo "================================================"
echo ""
go tool cover -func=coverage.out | tail -1
echo ""

echo "📈 Generating HTML coverage report..."
go tool cover -html=coverage.out -o coverage.html

echo "✅ Coverage report generated: coverage.html"
echo ""
echo "To view the HTML report, run:"
echo "  open coverage.html"
