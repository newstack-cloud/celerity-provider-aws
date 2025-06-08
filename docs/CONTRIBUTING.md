# Contributing to Celerity Provider for AWS

## Setup

Ensure git uses the custom directory for git hooks so the pre-commit and commit-msg linting hooks
kick in.

```bash
git config core.hooksPath .githooks
```

### NPM dependencies

There are npm dependencies that provide tools that are used in git hooks and scripting for the provider.

Install dependencies from the root directory by simply running:
```bash
yarn
```

## Further documentation

- [Commit Guidelines](./COMMIT_GUIDELINES.md)
- [Source Control and Release Strategy](./SOURCE_CONTROL_RELEASE_STRATEGY.md)
