# Zen CLI GPG Signing Keys

This directory contains the GPG public keys used to sign Zen CLI releases.

## Current Release Signing Key

**File:** `zen-signing-key.gpg`

**Key Details:**
- **Key ID:** `C4F5A887D4E02E41`
- **Fingerprint:** `0627E96AC3B78B66847CC2F8C4F5A887D4E02E41`
- **Email:** releases@daddia.com
- **Expires:** 2027-11-20

## Import Key

```bash
curl -fsSL https://raw.githubusercontent.com/daddia/zen-cli/main/keys/zen-signing-key.gpg | gpg --import
```

## Verify Release Signature

```bash
# Download release and signature
wget https://github.com/daddia/zen-cli/releases/download/v1.0.0/checksums.txt
wget https://github.com/daddia/zen-cli/releases/download/v1.0.0/checksums.txt.sig

# Verify
gpg --verify checksums.txt.sig checksums.txt
```

**Expected output:**
```
gpg: Good signature from "Zen CLI Release Signing <releases@daddia.com>"
```

## Automatic Verification

The install scripts automatically verify checksums:
```bash
curl -fsSL https://zen.daddia.com/cli/install | bash
```

## More Information

See [docs/security/README.md](../docs/security/README.md) for complete security documentation.

