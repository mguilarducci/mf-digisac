---
name: changelog
description: >
  Generate or update the CHANGELOG.md for a new release version.
  Use when the user says "generate changelog", "update changelog", "write release notes",
  or asks to prepare a changelog for a version like "changelog for 0.5.3".
---

<!-- Adapted from entireio/cli's .claude/skills/changelog (MIT, see NOTICE). -->

# Changelog Generation

Generate a changelog entry for a new release by analyzing merged PRs since the last tag.
The release workflow (`.github/workflows/release.yml`) publishes this entry as the
GitHub release notes and fails if it is missing, so run this before `mise run release vX.Y.Z`.

## Parameters

The user provides:
- **Version number** -- e.g., `0.5.3`
- **Additional PRs** -- optionally, PRs not yet merged that should be included

## Step 1: Gather Data

1. Find the previous release tag: `git tag --sort=-version:refname | head -1`
   (no tag yet: this is the first release, so use every merged PR)
2. List all merge commits since that tag: `git log <tag>..HEAD --merges --format='%s' | grep -v "Merge branch"`
3. Get PR details with: `gh pr list --state merged --base main --search "merged:><tag-date>" --json number,title,author,body --limit 50`
4. List contributors: `git log <tag>..HEAD --format='%an' --no-merges | sort | uniq -c | sort -rn`

## Step 2: Identify External Contributors

The repository owner (`mguilarducci`) is internal. Everyone else is external. Only list
external contributors in the Thanks section. Omit it when there are none.

## Step 3: Write the Entry

Insert the new version section at the top of `CHANGELOG.md`, after the header and before the previous version.

### Format

Follow the existing style in CHANGELOG.md exactly:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added

- Feature description with context ([#123](https://github.com/mguilarducci/mf-digisac/pull/123))

### Changed

- Change description ([#456](https://github.com/mguilarducci/mf-digisac/pull/456))

### Fixed

- Bug fix description ([#789](https://github.com/mguilarducci/mf-digisac/pull/789))

### Housekeeping

- Maintenance item ([#101](https://github.com/mguilarducci/mf-digisac/pull/101))

### Thanks

Thanks to @contributor for description of contribution!
```

### Style Rules

- **Sections**: Added, Changed, Fixed, Housekeeping, Thanks. Omit empty sections.
- **Each bullet**: starts with a dash, concise description, PR link(s) at the end
- **Group related PRs** into a single bullet when they're part of the same feature/fix
- **Work-in-progress features**: call out explicitly, e.g., "Feature X (work in progress): ..."
- **Known limitations**: note inline, e.g., "Note: plugin upgrades are not yet supported"
- **Thanks section**: only external contributors. Name what they contributed specifically.
- **Dependency bumps**: group into a single Housekeeping bullet unless a bump fixes a notable bug
- **PR links**: always use full URL format `[#N](https://github.com/mguilarducci/mf-digisac/pull/N)`
- **No trailing period** on bullet items
- **Date**: use the current date in YYYY-MM-DD format

### Categorization Guide

- **Added**: new features, new commands, new plugin capabilities, new CI workflows
- **Changed**: behavior changes, flag or output changes, UX changes, migrations
- **Fixed**: bug fixes, platform-specific fixes
- **Housekeeping**: dependency bumps, docs, refactors, CI improvements, test improvements
