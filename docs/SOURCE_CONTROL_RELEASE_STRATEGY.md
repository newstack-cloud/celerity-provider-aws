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

e.g. 0.1.0, 1.0.0-next.1
```

## Release workflow

1. Ensure all relevant changes have been merged (rebased) into the trunk (main). The release-please GitHub actions workflow will maintain a release PR that will be updated with the latest changes based on the conventional commit messages.
2. Ensure the version in `main.go` is updated to the next version number indicated in the release PR.
3. Review the release notes and change log changes in the release PR, update the release notes as necessary.
4. Once the release notes are ready, merge the release PR into main.
5. The release-please GitHub actions workflow will create a release tag and a draft release. The creation of the tag will trigger the release publishing workflow.
6. The release publishing workflow will build all the artifacts for the provider, generate a `docs.json` file for the plugin (to be consumed by the Celerity Registry) and publish the release or pre-release in GitHub.

## Pre-releases

When you want to create a pre-release, you should set `prerelease` to `true` in the release-please-config.json file.

This will cause the release-please GitHub actions workflow to create a pre-release version of the provider using the `-next.N` suffix.

Once the current pre-release version of the package is deemed stable, you should remove `prerelease` from the release-please-config.json file.
