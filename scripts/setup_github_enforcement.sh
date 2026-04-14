#!/usr/bin/env bash

set -euo pipefail

REPO_NAME="${1:-nexus}"
VISIBILITY="${2:-private}" # private|public
DEFAULT_BRANCH="main"
REQUIRED_CHECK="backend-go-tests"

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI is required."
  exit 1
fi

if ! gh auth status >/dev/null 2>&1; then
  echo "GitHub CLI is not authenticated. Run: gh auth login"
  exit 1
fi

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "Run this script inside a git repository."
  exit 1
fi

if ! git remote get-url origin >/dev/null 2>&1; then
  echo "No origin remote found. Creating GitHub repository ${REPO_NAME} (${VISIBILITY})..."
  gh repo create "${REPO_NAME}" --"${VISIBILITY}" --source=. --remote=origin --push
else
  echo "Origin remote already exists. Pushing current branch..."
  git push -u origin "${DEFAULT_BRANCH}"
fi

OWNER_REPO="$(gh repo view --json nameWithOwner --jq .nameWithOwner)"
echo "Configuring branch protection on ${OWNER_REPO}:${DEFAULT_BRANCH}"
set +e
PROTECTION_OUTPUT="$(
gh api --method PUT "repos/${OWNER_REPO}/branches/${DEFAULT_BRANCH}/protection" \
  --input - <<EOF
{
  "required_status_checks": {
    "strict": true,
    "contexts": ["${REQUIRED_CHECK}"]
  },
  "enforce_admins": true,
  "required_pull_request_reviews": {
    "required_approving_review_count": 1,
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": false
  },
  "restrictions": null,
  "required_linear_history": true,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "block_creations": false,
  "required_conversation_resolution": true,
  "lock_branch": false,
  "allow_fork_syncing": true
}
EOF
)"
PROTECTION_EXIT_CODE=$?
set -e

if [ ${PROTECTION_EXIT_CODE} -ne 0 ]; then
  echo "Could not apply branch protection automatically."
  echo "${PROTECTION_OUTPUT}"
  echo
  echo "Common cause: private repo in a plan that does not include branch protection."
  echo "Options:"
  echo "  1) Upgrade GitHub plan with branch protection support for private repos"
  echo "  2) Change repository visibility to public and rerun this script"
  exit ${PROTECTION_EXIT_CODE}
fi

echo "Branch protection configured successfully."
echo "Required check: ${REQUIRED_CHECK}"
