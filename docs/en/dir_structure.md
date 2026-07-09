# Full Directory Structure

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:09:10.824Z pushedAt=2026-06-09T01:11:14.879Z -->

The complete directory structure of the project is as follows:

```text
mef                                # Project root directory
├── build                          # Build-related directory
├── docs                           # Documentation directory
│   └── zh                         # Chinese document directory
└── src                            # Source code directory
    ├── common-utils               # Common utility library
    │   ├── backuputils            # Backup tool
    │   ├── build                  # Build-related directory
    │   ├── cache                  # Cache management
    │   ├── checker                # Verification tool
    │   ├── cmsverify              # CMS verification tool
    │   ├── database               # Database operation tool
    │   ├── envutils               # Environment variable tool
    │   ├── fileutils              # File operation tool
    │   ├── httpsmgr               # HTTPS management tool
    │   ├── hwlog                  # Logging tool
    │   ├── k8stool                # Kubernetes tool
    │   ├── kmc                    # KMC tool
    │   ├── limiter                # Rate limiter
    │   ├── logmgmt                # Log management
    │   ├── modulemgr              # Module manager
    │   ├── rand                   # Random number tool
    │   ├── terminal               # Terminal tool
    │   ├── test                   # Test directory
    │   ├── tls                    # TLS tool
    │   ├── utils                  # General tool
    │   ├── websocketmgr           # WebSocket manager
    │   ├── x509                   # X.509 tool
    │   └── xcrypto                # Encryption tool
    ├── device-plugin              # Device plugin
    │   ├── build                  # Build directory
    │   ├── doc                    # Documentation
    │   └── pkg                    # Main program code
    ├── mef-center                 # MEFCenter core code
    │   ├── alarm-manager          # Alarm management
    │   ├── build                  # Build configuration
    │   ├── cert-manager           # Certificate management
    │   ├── common                 # Common module
    │   ├── edge-manager           # Edge manager
    │   ├── mef-center-install     # MEF installation tool
    │   ├── nginx-manager          # Nginx manager
    │   ├── opensource             # Open-source component directory
    │   └── platform               # Platform module directory
    ├── mef-edge                   # MEFEdge code
    │   ├── build                  # Build configuration
    │   └── edge-installer         # Edge component directory
    │       ├── build              # Build configuration
    │       ├── cmd                # Main program entry
    │       ├── config             # Configuration directory
    │       ├── pkg                # Main program code
    │       ├── script             # Script directory
    │       └── tool               # Tool directory
```
