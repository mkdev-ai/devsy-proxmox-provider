# Spec Delta

## Purpose

Defines the observable contract for the Devsy Proxmox provider workspace so that
the Devsy daemon installs with an absolute state root and the Docker CLI can
authenticate to registries through `docker-credential-devsy`, allowing a devcontainer
to be prepared and started.

## ADDED Requirements

### Requirement: Devsy daemon state root must be absolute

The Devsy daemon installer SHALL expand a leading `~` to the real home directory
(`os.UserHomeDir`) and SHALL require the resolved daemon state root to be an
absolute path. It SHALL NOT write to a relative location or to a non-expanded `~`.

#### Scenario: state root resolves to an absolute path from $HOME

- **WHEN** the Devsy daemon installer chooses its default state root and `$HOME`
  is set to a valid directory
- **THEN** the root is `$HOME/.devsy/data` resolved to an absolute path and daemon
  installation proceeds

#### Scenario: unexpanded tilde is rejected

- **WHEN** the daemon installer resolves a state root equal to the literal string
  `~/.devsy/data` without expanding it
- **THEN** it exits with an error naming the bad root and the required absolute
  form, and does not attempt to write to it

### Requirement: docker-credential-devsy helper is discoverable and invocable

The provider workspace SHALL make `docker-credential-devsy` available on
`$PATH` and SHALL guarantee the Docker CLI can invoke it when resolving
credentials for a registry.

#### Scenario: helper is present on PATH

- **WHEN** provider workspace setup completes
- **THEN** the `docker-credential-devsy` executable is installed in a
  PATH-accessible directory and `command -v docker-credential-devsy` locates it

#### Scenario: Docker invokes the helper

- **WHEN** the Docker CLI resolves registry credentials
- **THEN** `docker-credential-devsy get` returns the credential map for the
  requested registry host, and devcontainer prepare/start proceeds

### Requirement: devcontainer prepare and start succeed end to end

A created workspace SHALL be started so that the devcontainer is prepared and
running, subject to the two requirements above.

#### Scenario: full workspace start

- **WHEN** a Devsy workspace is created and launched
- **THEN** the Devsy daemon installs with an absolute state root, the VM is
  started, an IP is resolved, the agent connects, the devcontainer prepares, and
  the container starts

## Notes

- This change is environment/credential-setup focused; it does not amend the
  Proxmox provisioning or lifecycle logic.
