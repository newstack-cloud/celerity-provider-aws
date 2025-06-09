# Contributing to Celerity Provider for AWS

## Setup

Ensure git uses the custom directory for git hooks so the pre-commit and commit-msg linting hooks
kick in.

```bash
git config core.hooksPath .githooks
```

### Prerequisites

- [Go](https://golang.org/dl/) >=1.23
- [GolangCI-Lint](https://golangci-lint.run/welcome/install/#local-installation) - Used for linting and formatting.
- [Node.js](https://nodejs.org/en/download/) - Used for running scripts for commit message linting.
- [Yarn](https://yarnpkg.com/getting-started/install) - Used for managing dependencies for commit message linting.

Dependencies are managed with Go modules (go.mod) and will be installed automatically when you first
run tests.

If you want to install dependencies manually you can run:

```bash
go mod download
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
