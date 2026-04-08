# Source Control & Release Strategy

## Source control & development workflow

- Development work by core contributors should be carried out on the main branch for most contributions, with the exception being longer projects (weeks or months worth of work) or experimental new versions of the provider. For the exceptions, feature/hotfix branches should be used.
- All development work by non-core contributors should be carried out on feature/hotfix branches on your fork, pull requests should be utilised for code reviews and merged (**rebase!**) back into the main branch of the primary repo.
- All commits should follow the [commit guidelines](./COMMIT_GUIDELINES.md).
- Work should be committed in small, specific commits where it makes sense to do so.

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
5. The release-please GitHub actions workflow will create a release tag and a **draft** release. The creation of the tag will trigger the release publishing workflow.
6. The release publishing workflow will build all the artifacts for the provider, generate a `docs.json` file for the plugin (to be consumed by the Bluelink Registry), upload all artifacts to the draft release, and then publish it.

### Why draft releases matter

The release-please config has `"draft": true` set intentionally. This is critical for correct integration with the Bluelink Registry.

The Bluelink Registry uses a GitHub webhook that listens for `release` events with the `published` action. When a release is published, the registry downloads the release artifacts (`bluelink-registry-info.json`, `docs.json`, platform binaries, checksums, and GPG signatures) and processes them to update the registry.

If release-please were to create a **published** (non-draft) release directly, the `published` webhook event would fire immediately — before goreleaser has had a chance to build and upload the artifacts. The registry would receive the webhook, attempt to download the artifacts, and fail because they don't exist yet.

The draft release approach solves this race condition:

```
1. release-please creates a draft release + tag
   → No "published" event fired
   → Registry is not notified

2. Tag push triggers the release workflow
   → goreleaser builds binaries for all platforms
   → plugin-docgen generates docs.json
   → All artifacts are uploaded to the draft release

3. Final workflow step publishes the draft (gh release edit --draft=false)
   → "published" event fires NOW
   → Registry webhook receives the event
   → All artifacts are available for download
```

**Do not remove `"draft": true` from `release-please-config.json`**.

### Release artifacts

The following artifacts are produced by goreleaser and included in each release:

| Artifact | Naming Convention | Purpose |
|----------|-------------------|---------|
| Platform binaries | `bluelink-provider-aws_{version}_{os}_{arch}.zip` | Plugin binaries for each supported platform |
| Registry info | `bluelink-provider-aws_{version}_registry_info.json` | Metadata consumed by the Bluelink Registry (protocols, dependencies, UI config) |
| Docs | `bluelink-provider-aws_{version}_docs.json` | Generated plugin documentation (display name, resource/data source/link docs) |
| Checksums | `bluelink-provider-aws_{version}_SHA256SUMS` | SHA-256 checksums for all release files |
| GPG signature | `bluelink-provider-aws_{version}_SHA256SUMS.sig` | GPG signature of the checksum file |

The `bluelink-registry-info.json` and `docs.json` source files are generated during the release workflow and included as extra files in the goreleaser config.

### Bluelink Registry integration

The Bluelink Registry at [registry.bluelink.dev](https://registry.bluelink.dev) automatically discovers and publishes new versions of the provider plugin.

**How it works:**

1. When a plugin is first registered with the registry, a GitHub webhook is dynamically created on the provider repository. The webhook is configured to listen for `release` events and is secured with an HMAC-SHA1 secret unique to the organisation.
2. When a release is published, GitHub sends a `POST` request to the registry's webhook endpoint (`/webhook/gh/{organisation}`).
3. The registry validates the webhook signature, checks that the event action is `published`, and extracts the release information.
4. The registry identifies the plugin type and ID from the repository name (e.g., `bluelink-provider-aws` becomes provider `newstack-cloud/aws`).
5. Release artifacts are downloaded and processed — registry info for metadata, docs for documentation, and platform binaries are catalogued with their checksums.

**Requirements for the provider repository:**

- The repository name must follow the pattern `bluelink-provider-{name}` for the registry to recognise it as a provider plugin.
- Release tags must use semantic versioning with a `v` prefix (e.g., `v0.1.0`, `v1.0.0-next.1`).
- The `bluelink-registry-info.json` file must be present in the release with the required fields (`supportedProtocols`, `dependencies`, and optionally `ui.referencedLinkPlugins`).
- The `docs.json` file must be present with at least a `displayName` field.
- The release must not be a draft when the registry processes it — all artifacts must be available at the time the `published` event is received.

## Pre-releases

When you want to create a pre-release, you should set `prerelease` to `true` in the release-please-config.json file.

This will cause the release-please GitHub actions workflow to create a pre-release version of the provider using the `-next.N` suffix.

Once the current pre-release version of the package is deemed stable, you should remove `prerelease` from the release-please-config.json file.
