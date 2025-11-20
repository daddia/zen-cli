# Zen CLI Security

## Release Signing

All Zen CLI releases are signed with GPG for verification and security.

### GPG Public Key

**Key ID:** `C4F5A887D4E02E41`  
**Fingerprint:** `0627E96AC3B78B66847CC2F8C4F5A887D4E02E41`  
**Email:** releases@daddia.com  
**Expires:** 2027-11-20

### Verify Release Signatures

**Import the public key:**
```bash
curl -fsSL https://raw.githubusercontent.com/daddia/zen-cli/main/keys/zen-signing-key.gpg | gpg --import
```

**Verify a release:**
```bash
# Download release and signature
wget https://github.com/daddia/zen-cli/releases/download/v1.0.0/checksums.txt
wget https://github.com/daddia/zen-cli/releases/download/v1.0.0/checksums.txt.sig

# Verify signature
gpg --verify checksums.txt.sig checksums.txt
```

**Expected output:**
```
gpg: Signature made ...
gpg: Good signature from "Zen CLI Release Signing <releases@daddia.com>"
```

### SHA256 Checksums

All release binaries include SHA256 checksums in `checksums.txt`.

**Verify a binary:**
```bash
# Download binary and checksum
wget https://github.com/daddia/zen-cli/releases/download/v1.0.0/zen_1.0.0_linux_amd64.tar.gz
wget https://github.com/daddia/zen-cli/releases/download/v1.0.0/checksums.txt

# Verify (Linux)
sha256sum -c checksums.txt --ignore-missing

# Verify (macOS)
shasum -a 256 -c checksums.txt
```

### Security Policy

**Reporting Vulnerabilities:**
- Email: security@daddia.com
- GitHub Security Advisories: https://github.com/daddia/zen-cli/security/advisories

**Response Time:**
- Critical: 24 hours
- High: 72 hours
- Medium/Low: 1 week

### Automated Verification

The install scripts automatically verify checksums:

```bash
curl -fsSL https://zen.daddia.com/cli/install | bash
```

This script:
1. Downloads binary from GitHub releases
2. Downloads checksum file
3. Verifies SHA256 checksum
4. Only installs if checksum matches

No manual verification needed for standard installations.

