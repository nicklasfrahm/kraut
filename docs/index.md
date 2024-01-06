# Kraut

`kraut` is an infrastructure orchestrator built on top of Kubernetes. The goal is to provide APIs to build private clouds using Kubernetes as a control plane and API.

## Overview

- [**Management**](./management.md)  
  As `kraut` is built on top of **bare-metal infrastructure**, it requires foundational management APIs to connect to existing infrastructure, such as network appliances or bare-metal servers.

- [**Networking**](./networking.md)  
  `kraut` provides low-level APIs to manage network infrastructure, such as an `Interface` or a `Network`.

If you are consuming the APIs you may also find the [auto-generated CRD documentation][crd-docs] useful.

## Architectural principles

- **Agentless**  
  `kraut` aims to be agentless, meaning that it does not require any agent to be installed on the managed infrastructure. This may mean an increased amount of management traffic but also avoids the need to install and maintain agents, possibly allowing for easier development of new appliance drivers.

## License

`kraut` is and will always be licensed under the terms of the MIT license.

[crd-docs]: https://doc.crds.dev/github.com/nicklasfrahm/kraut
