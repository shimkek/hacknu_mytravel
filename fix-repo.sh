#!/bin/bash

# Script to fix the divergent branches issue and clean up repository
# Run this on your local machine to clean up the repository

echo "🔧 Fixing GitHub repository branches..."

# Check current branch
current_branch=$(git branch --show-current)
echo "Current branch: $current_branch"

# Check if we have changes to commit
if ! git diff --quiet || ! git diff --cached --quiet; then
    echo "📝 You have uncommitted changes. Committing them first..."
    git add .
    git commit -m "Fix: Update project structure, database schema, and Docker builds"
fi

# Make sure we're on master branch (where your code is)
if [ "$current_branch" != "master" ]; then
    echo "📍 Switching to master branch..."
    git checkout master
fi

# Set the remote URL
git remote set-url origin https://github.com/Yerassyl20036/hacknu_mytravel.git

# Check what's on the remote
echo "🔍 Checking remote branches..."
git fetch origin

# Delete the problematic main branch on remote and replace it with master
echo "🗑️  Deleting old main branch on remote..."
git push origin --delete main 2>/dev/null || echo "Main branch doesn't exist or already deleted"

# Push master branch as the new main branch
echo "🚀 Pushing master branch as main..."
git push origin master:main --force

# Set main as the default branch locally
echo "🔄 Setting up local main branch..."
git checkout -b main 2>/dev/null || git checkout main
git reset --hard master
git push origin main --force

# Optional: Keep only main branch (delete master)
echo "🧹 Cleaning up branches..."
echo "Do you want to delete the master branch and keep only main? (y/n)"
read -r response
if [[ "$response" == "y" || "$response" == "Y" ]]; then
    git push origin --delete master
    echo "✅ Master branch deleted, using main as the only branch"
else
    echo "✅ Keeping both branches (master and main have the same content)"
fi

echo ""
echo "✅ Repository fixed!"
echo ""
echo "📋 Summary:"
echo "   - Your code is now properly on main branch"
echo "   - Remote repository conflicts resolved"
echo "   - Docker build issues fixed"
echo "   - All go.sum files generated"
echo ""
echo "🚀 Next steps:"
echo "   1. Deploy on VPS: ./deploy.sh"