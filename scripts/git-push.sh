#!/bin/bash

# Velure Repository - Git Automation Script
# Automatically stages, commits, and pushes all changes

set -e

echo "🚀 Velure - Git Automation"
echo "=========================="

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Error: Not in a git repository"
    exit 1
fi

# Show current status
echo "📊 Current Git Status:"
git status --short

# Check if there are any changes
if [[ -z $(git status --porcelain) ]]; then
    echo "✅ No changes to commit"
    exit 0
fi

# Stage all changes
echo ""
echo "📦 Staging all changes..."
git add .

# Show what will be committed
echo ""
echo "📝 Changes to be committed:"
git diff --cached --name-status

# Create commit message
echo ""
if [ -n "$1" ]; then
    COMMIT_MSG="$1"
else
    COMMIT_MSG="chore: repository reorganization and infrastructure updates

- Reorganized services into services/ directory
- Moved infrastructure to infrastructure/ directory  
- Updated documentation structure in docs/
- Enhanced CI/CD pipelines for new structure
- Updated Terraform configurations
- Improved Makefile with comprehensive automation
- Updated Docker Compose for new paths
- Added security and quality workflows
- Consolidated and improved documentation"
fi

echo "💬 Commit message:"
echo "\"$COMMIT_MSG\""

# Confirm before committing
echo ""
read -p "🤔 Proceed with commit and push? (y/N): " -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
    # Commit changes
    echo "✅ Committing changes..."
    git commit -m "$COMMIT_MSG"
    
    # Check current branch
    CURRENT_BRANCH=$(git branch --show-current)
    echo "📍 Current branch: $CURRENT_BRANCH"
    
    # Push to remote
    echo "🚢 Pushing to origin/$CURRENT_BRANCH..."
    git push origin "$CURRENT_BRANCH"
    
    echo ""
    echo "🎉 Successfully pushed all changes!"
    echo "🔗 Repository: https://github.com/icl00ud/velure"
    
    # Show final status
    echo ""
    echo "📊 Final status:"
    git status
else
    echo "❌ Aborted"
    exit 1
fi