# Release signing

The release identity established on 2026-09-12 uses the public key in `keys/profitctl-release-cosign.pub`. Its SHA-256 file fingerprint is `aebe6076c728012a09e7795ac9c76cfb1729db640bdc798932c00bd6b3c3023f`.

The owner approved creating this identity because the previous private key was unavailable. The previous public key remains in Git history for historical verification; replacing it does not re-sign or authenticate older assets.

## Local custody

The encrypted Sigstore private key and its random password are stored together as one macOS Keychain generic-password item:

- Service: `co.intelip.profitctl.release-signing`
- Account: `release-v1`
- Local helper: `~/.local/libexec/profitctl-release-keychain`
- Helper source: `~/.local/libexec/profitctl-release-keychain.swift`

The helper uses the native Security framework. It exposes metadata checks, public-key recovery, and signing; it has no command that exports the private key or password. Secret values reach Cosign only through child-process environment variables. Temporary encrypted key-generation files are removed after storage. No private credential belongs in this repository or release attachments.

```sh
~/.local/libexec/profitctl-release-keychain status
~/.local/libexec/profitctl-release-keychain sign \
  dist/v0.4.0-rc.1 keys/profitctl-release-cosign.pub
bash scripts/release/verify-release-assets.sh \
  v0.4.0-rc.1 dist/v0.4.0-rc.1 keys/profitctl-release-cosign.pub
```

The local helper signs and verifies native archives, the Bun tarball, SPDX SBOMs, `SHA256SUMS`, and `RELEASE-SHA256SUMS`. Signatures are detached; transparency-log uploads are disabled. Check `RELEASE-SHA256SUMS` as well as signatures when reviewing the complete set.

The existing environment-based release script remains available for an explicitly configured signing environment. It requires `COSIGN_PRIVATE_KEY` and `COSIGN_PASSWORD`; do not copy these into shell startup files or command arguments. CI signing credentials have not been provisioned.

## Recovery and distribution

This setup provides local Keychain custody. No off-device backup or secret-manager replica has been created. The public key cannot recover the private key. If the Keychain item is lost, a new identity and an explicit trust-key update are required.

Distribute the public key through an independently trusted channel or pin its fingerprint before verification. A signature verifies possession of the signing key and asset integrity; it does not establish that the software is safe or that a release has been published. Signing does not authorize publication or deployment.
