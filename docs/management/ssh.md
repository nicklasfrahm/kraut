# SSH

This section describes how to enroll a `Host` using SSH.

## Prerequisites

Managing an host using SSH requires that the login shell of the user is `bash`.

### NX-OS

To enable the `bash` login shell on NX-OS, you need to run the following commands.

```shell
# Enter configuration mode.
configure terminal
# Configure bash as the login shell for the user "admin".
username admin shelltype bash
# Exit configuration mode.
exit
# Persist settings between reboots.
copy running-config startup-config
```

After reconnecting to the host, `bash` is now your default shell. You may enter the NX-OS VSH for configuration commands by running `vsh`.

## Configuration

### Secret

Below you may find an example of a `Secret` for an SSH connection using all possible keys. The usage of an encrypted private key and a jump host is optional.

```yaml title="ssh-secret.yaml"
apiVersion: v1
kind: Secret
metadata:
  name: ssh-credentials
type: Opaque
stringData:
  # Passphrase if the private key is encrypted.
  passphrase: dont-check-this-into-your-repo-please
  # Password for authentication.
  # This is only used if the private key is not provided.
  passwordInsecure: possible-but-not-recommended
  # Private key for authentication. This takes precedence over the password.
  key: |
    -----BEGIN OPENSSH PRIVATE KEY-----
    REDACTED
    -----END OPENSSH PRIVATE KEY-----
  # Passphrase for an encrypted private key while using a jump host.
  proxyPassphrase: dont-check-this-into-your-repo-please
  # Password for authentication while using a jump host.
  # This is only used if the private key is not provided.
  proxyPasswordInsecure: possible-but-not-recommended
  # Private key for authentication while using a jump host.
  # This takes precedence over the password.
  proxyKey: |
    -----BEGIN OPENSSH PRIVATE KEY-----
    REDACTED
    -----END OPENSSH PRIVATE KEY-----
```

### Host

Below, you may find a simple example where the controller will connect directly to the host.

```yaml title="alfa.yaml"
TODO: Add YAML example.
```

In more complex setup, you many need to use a jump host to connect to your host. This can be done as well.

```yaml title="distswitch00.yaml"
TODO: Add YAML example.
```
