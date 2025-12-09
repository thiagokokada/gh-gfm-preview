# Reproducibility

This document describes how anyone can independently reproduce the builds and
assets of `gh-gfm-preview` from source, and verify that a given release
corresponds exactly to that source.

## Why reproducibility matters

- Ensures that pre-built binaries released via GitHub are indeed built from the
  publicly available source.
- Helps detect tampering, build non-determinism or missing dependencies.
- Enables contributors and downstream packagers (e.g. distributions) to
  reproduce builds in a controlled environment.

## What `gh-gfm-preview` provides to support reproducibility

### Source

The entire codebase (Go code, assets, configuration, build scripts,
dependencies) is stored in this repository, under version control.

### Third party assets

Static assets that comes from third-party sources are downloaded and verified
using [generate-assets.go](internal/server/_tools/generate-assets.go) script.
It means you can delete `internal/server/static/generated` directory, rerun `go
generate ./...` and should get no differences between the files. You can also
audit the script to make sure the files are coming from their respective
sources.

### Dependency management

The project tracks Go module dependencies via `go.mod` and `go.sum`, which
ensures that deterministic dependency versions are used.

## Releases

The [releases][1] provided from this repository includes [build
attestations][2]. Build attestations plus the fact the source is reproducible
significantly strengthen supply-chain security by enabling users to:

- Verify that releases are built from the exact commit they claim.
  + If the published binary doesn't match the source, the attestation will not
    validate.
- Detect tampering or compromised CI environments.
  +  If someone injects malicious code into the build step or uploads a fake
     binary, the attestation signature will fail verification.
- Ensure that builds were produced in a trusted execution environment.
  + GitHub’s signed provenance ensures the binary really came from GitHub
    Actions, not a local machine or unknown builder.
- Establish immutable evidence of the build process.
  +  GitHub submits signatures to transparency logs, providing non-repudiation.
- Enable downstream distributors to audit releases.
  + They can reproduce the build locally, compare artifacts, and verify
    provenance via Sigstore tooling.

The attestation can be validated with [gh][3] tool, for example:

```console
$ curl -sL https://github.com/thiagokokada/gh-gfm-preview/releases/download/v0.9.4/linux-arm64 -o gh-gfm-preview
$ gh attestation verify gh-gfm-preview --owner thiagokokada
Loaded digest sha256:8fbd38807258c45a631f543b9a1baf5b3bf090ae4ccd6279a4571653abf228d9 for file://gh-gfm-preview
Loaded 1 attestation from GitHub API

The following policy criteria will be enforced:
- Predicate type must match:................ https://slsa.dev/provenance/v1
- Source Repository Owner URI must match:... https://github.com/thiagokokada
- Subject Alternative Name must match regex: (?i)^https://github.com/thiagokokada/
- OIDC Issuer must match:................... https://token.actions.githubusercontent.com

✓ Verification succeeded!

The following 1 attestation matched the policy criteria

- Attestation #1
  - Build repo:..... thiagokokada/gh-gfm-preview
  - Build workflow:. .github/workflows/release.yml@refs/tags/v0.9.4
  - Signer repo:.... thiagokokada/gh-gfm-preview
  - Signer workflow: .github/workflows/release.yml@refs/tags/v0.9.4
```

[1]: https://github.com/thiagokokada/gh-gfm-preview/releases/
[2]: https://docs.github.com/en/actions/how-tos/secure-your-work/use-artifact-attestations/use-artifact-attestations
[3]: https://cli.github.com/
