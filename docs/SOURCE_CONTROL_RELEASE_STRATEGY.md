# Source Control & Release Strategy

## Source control & development workflow

- Development work by core contributes should be carried out on the main branch for most contributions, with the exception being longer projects (weeks or months worth of work) or experimental new versions of the provider. For the exceptions, feature/hotfix branches should be used.
- All development work by non-core contributes should be carried out on feature/hotfix branches on your fork, pull requests should be utilised for code reviews and merged (**rebase!**) back into the main branch of the primary repo.
- All commits should follow the [commit guidelines](./COMMIT_GUIDELINES.md).
- Work should be commited in small, specific commits where it makes sense to do so.

## Release strategy

Tags used for releases need to be in the following format:

```
MAJOR.MINOR.PATCH(-PRE_RELEASE_SUFFIX)?

e.g. 0.1.0, 0.1.0-alpha.1, 0.1.0-beta.1, 0.1.0-rc.1
```

## Release workflow

1. Ensure all relevant changes have been merged (rebased) into the trunk (main).
2. Create a new release branch for `release/MAJOR.MINOR.PATCH(-PRE_RELEASE_SUFFIX)?` (e.g. `release/0.1.1`) with the approximate next version. (This branch is short-lived so it is not crucial to get the version 100% correct)
3. Push the release branch and this will trigger a GitHub actions workflow that will determine the actual version from commits and update the change log for the target application or library.
4. The automated workflow from step 3 will create a PR that generates a preliminary set of release notes. Review and edit the release notes accordingly and then rebase the PR into main. (These release notes will be used in a further automated release publishing step)
5. Rebasing the PR into main will trigger the process of creating the tag and release in GitHub along with building and publishing artifacts for the provider.
