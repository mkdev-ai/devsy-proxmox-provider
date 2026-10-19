# Spec Delta

## Purpose

Defines the observable contract for the Devsy Proxmox provider workspace so that
the Devsy daemon installs with an absolute state root and the Docker CLI can
authenticate to registries through `docker-credential-devsy`, allowing a devcontainer
to be prepared and started.

## ADDED Requirements

### Requirement: Devsy agent dataPath must be absolute in the provider manifest

The provider SHALL NOT pin a literal `~` in the agent `dataPath` in its
provider manifest. Devsy's daemon installer rejects a non-expanded `~`
as a non-absolute state root and aborts workspace creation.

#### Scenario: manifest omits dataPath

- **WHEN** a user adds the provider via Devsy
- **THEN** the manifest does not contain `dataPath: ~` and Devsy uses its own
  absolute default state root

#### Scenario: manifest dataPath explicit

- **WHEN** the manifest sets agent dataPath
- **THEN** it is an absolute path (or absent), and a literal `~` is rejected

### Requirement: Daemon installer never receives a non-absolute state root

The provider SHALL configure Devsy so the daemon installer never sees a literal
`~` or a relative path as the daemon state root.

#### Scenario: devsy up creates a workspace

- **WHEN** `devsy up` is run for a workspace
- **THEN** the Devsy daemon installs successfully and the provider reports the
  workspace as ready, without emitting `state root must be absolute`.

### Requirement: Document daemon data root override

The provider documentation SHALL document how to override the Devsy daemon data
root for a user (`DEVSY_HOME` / `--home`), rather than requiring the provider
manifest to hardcode a per-user path.

#### Scenario: user configures a custom daemon data root

- **WHEN** a user needs a custom daemon data root on their host
- **THEN** they set `DEVSY_HOME` or pass `--home=PATH`; the provider manifests
  remain correct and do not hardcoding a guest-user home directory.
