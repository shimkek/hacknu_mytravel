#!/bin/bash

# Script to fix the divergent branches issue
# Run this on your local machine to clean up the repository

echo "🔧 Fixing GitHub repository branches..."

# Check current branch
current_branch=$(git branch --show-current)
echo "Current branch: $current_branch"

# Check if we have changes to commit
if ! git diff --quiet || ! git diff --cached --quiet; then
    echo "📝 You have uncommitted changes. Committing them first..."
    git add .
    git commit -m "Fix: Update project structure and database schema"
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

# Force push master branch to override main
echo "🚀 Pushing master branch to override main..."
git push origin master:main --force

# Set main as the default branch and delete old master on remote
echo "🔄 Setting main as default branch..."
git push origin master:main --force

# Optional: Delete the old master branch on remote if you want main as default
# Uncomment the next line if you want to keep only main branch
# git push origin --delete master

echo "✅ Repository fixed!"
echo ""
echo "📋 Summary:"
echo "   - Your code from master branch is now on main branch"
echo "   - Remote repository has been updated"
echo "   - You can now deploy with ./deploy.sh"
echo ""
echo "🚀 Next steps:"
echo "   1. Run this script: ./fix-repo.sh"
echo "   2. Deploy on VPS: ./deploy.sh"