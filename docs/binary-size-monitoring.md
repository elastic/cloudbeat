# Binary Size Monitoring

This repository includes an automated binary size monitoring system that tracks changes to the cloudbeat binary size in pull requests.

## How it Works

The binary size monitor workflow (`.github/workflows/binary-size-monitor.yml`) automatically:

1. **Builds binaries** from both the PR branch and main branch
2. **Compares sizes** and calculates the percentage change
3. **Comments on PRs** with detailed size information
4. **Fails the workflow** if size increase exceeds the configured threshold

## Configuration

### Size Threshold

The workflow uses a configurable threshold to determine when a size increase is significant:

```yaml
env:
  # Size increase threshold percentage (e.g., 10 = 10% increase)
  SIZE_THRESHOLD: 10
```

You can modify this value to adjust sensitivity:
- Lower values (e.g., 5) will catch smaller size increases
- Higher values (e.g., 20) will only flag major size increases

### Customizing the Workflow

To modify the workflow behavior, edit `.github/workflows/binary-size-monitor.yml`:

- **Timeout**: Adjust `timeout-minutes` if builds take longer
- **Branches**: Modify `branches` array to target different branches
- **Build process**: Update build commands if build process changes

## PR Comments

The workflow automatically adds a sticky comment to PRs with a table showing:

| Branch | Size (MB) | Size (bytes) |
|--------|-----------|-------------|
| **PR Branch** | 45.67 MB | 47890432 bytes |
| **Main Branch** | 44.23 MB | 46389248 bytes |
| **Difference** | +1.44 MB | +1501184 bytes |

**Size Change:** +3.25%

✅ Binary size change is within acceptable limits.

## Workflow Failure

When the size increase exceeds the threshold, the workflow will:

1. **Add a warning comment** to the PR with ⚠️ indicator
2. **Fail the workflow** to prevent automatic merging
3. **Require manual review** of the size impact

## Troubleshooting

### Build Failures

The workflow includes fallback build methods:
1. First tries `mage build`
2. Falls back to direct `go build` if mage fails

### Size Calculation Issues

If you see errors in size calculation:
- Check that both binaries are successfully built
- Verify the `bc` utility is available in the runner
- Ensure proper file permissions on temporary files

## Bypassing the Check

In rare cases where a legitimate size increase needs to be merged:

1. **Document the reason** in the PR description
2. **Get approval** from code owners
3. **Temporarily increase** the threshold if needed
4. **Consider optimizations** to reduce size impact

## Integration with Other Workflows

This workflow is designed to complement existing CI workflows:
- Runs independently of unit tests and integration tests
- Uses the same hermit environment setup as other workflows
- Follows the same branch targeting patterns
- Respects the same concurrency controls