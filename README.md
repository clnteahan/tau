# Tau Package Manager
### Very in development package manager for any distribution.
## Goals (idealistic):
- Decentralized package management
  - Allows git sources
  - Allow ftp sources (Gnu, mirrors, etc.)
  - Remote http/https source (sourceforge, etc.)
- OS staging: automate the setup of an os via tau on another
- Variety of supported hardware and libraries
- Tools to translate recipes of other package managers
- Source and binary support
- Secure file transfer
- Automatic signature checking
- Hooks for other automation

## Build
```SH
go build -o -buildmode=shared build
```
or install with
```SH
go build -buildmode=exe -o /usr/bin
```