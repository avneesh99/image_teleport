# image_teleport

image_teleport is a CLI tool designed to push Docker images to remote EC2 instances, especially useful in environments with limited internet connectivity where pulling from DockerHub or AWS ECR is not possible.

## Features
- **Efficient Layer Synchronization:** Pushes only the modified Docker image layers, reducing transfer time and bandwidth usage.
- **Resumable Transfers:** Automatically resumes interrupted transfers, ensuring reliability in unstable network conditions.
- **Support for Bastion Hosts:** Allows pushing images through a bastion host for enhanced security in complex network setups.
- **Automatic Image Reconstruction:** Provides scripts to reconstruct and load Docker images on the remote host seamlessly.
- **Minimal Dependency on Internet:** Enables deployments to EC2 instances without direct internet access by avoiding the need to pull from Docker Hub or AWS ECR.

## Prerequisites

- Docker installed on the local machine
- rsync installed on both local and remote machines
- SSH access to the remote EC2 instance

## Installation

```
git clone https://github.com/avneesh99/image_teleport
go mod tidy
go build -o image_teleport
```

## Usage

```
image_teleport -image <image_name:tag> -remote <remote_host> -identity <path_to_ssh_key> [options]
```

## License
This project is licensed under the MIT License.


