# Automatic Version Tagging

This project uses GitHub Actions to automatically create version tags and releases when Pull Requests are merged into the main branch.

## How It Works

The automatic tagging workflow is defined in `.github/workflows/auto-tag.yml` and performs the following actions:

1. **Trigger**: The workflow is triggered when a Pull Request is closed (merged) into the main branch.

2. **Version Calculation**:

   - The workflow finds the latest version tag (format: `vX.Y.Z`)
   - It increments the patch version by 1 (e.g., `v1.0.0` → `v1.0.1`)
   - If no previous tag exists, it starts with `v1.0.0`

3. **Tag Creation**:

   - Creates an annotated Git tag with the new version
   - The tag message includes the PR number and title
   - Pushes the tag to the repository

4. **Release Creation**:
   - Automatically creates a GitHub Release using the new tag
   - The release includes:
     - The PR title and number
     - The PR description as the release notes
     - Information about who merged the PR

## Benefits

This automatic tagging system provides several benefits:

- **Consistent Versioning**: Ensures that each merged PR results in a properly versioned release
- **Traceability**: Links releases directly to the Pull Requests that created them
- **Documentation**: Automatically generates release notes from PR descriptions
- **CI/CD Integration**: The new tags can trigger other workflows, such as Docker image builds

## Best Practices for Pull Requests

To make the most of this automatic tagging system:

1. **Descriptive PR Titles**: Use clear, concise titles that describe the changes
2. **Detailed PR Descriptions**: Include comprehensive descriptions that can serve as good release notes
3. **One Feature Per PR**: Keep PRs focused on a single feature or fix for cleaner release notes

## Manual Version Bumps

For major or minor version bumps (instead of patch):

1. Create a tag manually before merging the PR:

   ```bash
   git tag -a v2.0.0 -m "Major version bump for XYZ feature"
   git push origin v2.0.0
   ```

2. The auto-tagging workflow will detect this as the latest tag and increment from there.

## Troubleshooting

If the automatic tagging doesn't work as expected:

1. Check the GitHub Actions logs for any errors
2. Ensure the PR was properly merged into the main branch
3. Verify that the repository has the correct permissions set for GitHub Actions

For more information on GitHub Actions, see [GitHub Actions Documentation](https://docs.github.com/en/actions).
