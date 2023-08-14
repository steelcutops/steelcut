# Steelcut

## Introduction

The "steelcut" library provides a comprehensive suite of functionalities to
manage Unix hosts, focusing on tasks related to SSH connections, package
management, file operations, system information retrieval, and host grouping.

## Functionalities

*SSH Connection Management*: Functions like Dial enable establishing SSH
connections with given network configurations, addresses, client config, and
timeouts.

*Host Management*: Various HostOptions, including setting the user, password, OS,
key passphrase, and SSH client for a Unix host. Functions like IsReachable,
ping, and sshable help to determine host accessibility.

*File Operations*: Abilities to copy files between local and remote paths, create
and delete directories, and set or get file permissions.

*Command Execution*: Execution of commands both locally and remotely. Options for
running commands with superuser privileges are also provided.

*Host Grouping*: Creation and management of host groups with functionalities like
running commands on all hosts in a group.

*Package Management*: Extensive package management functionalities, including
listing, adding, removing, and upgrading packages. Specific methods are provided
for different package managers like Yum, Apt, and Brew.

*System Information Retrieval*: Functions to retrieve CPU, memory, and disk usage,
along with listing running processes.

*Specialized Host Functions*: Specific functionalities for Linux and MacOS hosts,
including rebooting, shutting down, and checking updates.

*SSH Key Management*: Reading private keys from different sources, including the
SSH agent and the user's home directory.

*Unix Host Operations*: General operations on Unix hosts, like setting and getting
permissions, creating and deleting directories, etc.

## Summary

In summary, steelcut is a feature-rich library designed to ease the process of
managing and interacting with Unix hosts. Its diverse functionalities make it
suitable for a broad range of tasks, from basic file operations to complex
system management and orchestration.
