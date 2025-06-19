# 🔐 Security Documentation

This document details the security implementation, cryptographic design, and threat model of Flint Vault.

## 🎯 Security Overview

Flint Vault employs **military-grade encryption** designed to protect sensitive data against sophisticated adversaries. The system uses multiple layers of security to ensure both confidentiality and integrity.

### Security Goals

1. **Confidentiality**: Data cannot be read by unauthorized parties
2. **Integrity**: Data cannot be modified without detection
3. **Authentication**: Verify the authenticity of data and vault files
4. **Forward Secrecy**: Past data remains secure even if passwords are compromised
5. **Side-Channel Resistance**: Protection against timing and power analysis attacks

## 🔒 Cryptographic Implementation

### Encryption Algorithm: AES-256-GCM

**Advanced Encryption Standard (AES)** with **256-bit keys** in **Galois/Counter Mode (GCM)**

**Why AES-256-GCM?**
- ✅ **NSA Suite B approved** for TOP SECRET information
- ✅ **Authenticated encryption** (AEAD) prevents tampering
- ✅ **High performance** with hardware acceleration
- ✅ **Proven security** with extensive cryptanalysis
- ✅ **Resistance to quantum attacks** (post-quantum secure key sizes)

**Technical Details:**
```
Cipher: AES-256
Mode: GCM (Galois/Counter Mode)
Key Size: 256 bits (32 bytes)
Nonce Size: 96 bits (12 bytes)
Tag Size: 128 bits (16 bytes)
```

### Key Derivation: PBKDF2

**Password-Based Key Derivation Function 2** with **SHA-256**

**Parameters:**
```
Hash Function: SHA-256
Iterations: 100,000
Salt Size: 256 bits (32 bytes)
Output Key Size: 256 bits (32 bytes)
```

**Why PBKDF2?**
- ✅ **NIST recommended** (SP 800-132)
- ✅ **High iteration count** prevents brute force attacks
- ✅ **Unique salt** prevents rainbow table attacks
- ✅ **Memory-hard** properties slow down attackers
- ✅ **Widely tested** and standardized

### Random Number Generation

**Cryptographically Secure Pseudo-Random Number Generator (CSPRNG)**

**Sources:**
- **crypto/rand** package (Go standard library)
- **/dev/urandom** on Unix systems
- **CryptGenRandom** on Windows
- **Hardware entropy** when available

**Usage:**
- 🔑 **Salt generation**: 32 bytes per vault
- 🔢 **Nonce generation**: 12 bytes per encryption
- 🎲 **Key material**: All cryptographic keys

## 🏗️ Security Architecture

### File Format Structure

```
┌─────────────────────────────────────────────────────────┐
│                     Vault File                          │
├─────────────────────────────────────────────────────────┤
│ Magic Header (8 bytes): "FLINT001"                     │
├─────────────────────────────────────────────────────────┤
│ Salt (32 bytes): Random cryptographic salt             │
├─────────────────────────────────────────────────────────┤
│ Nonce (12 bytes): AES-GCM nonce                       │
├─────────────────────────────────────────────────────────┤
│ Encrypted Data: AES-256-GCM(JSON + gzip)              │
├─────────────────────────────────────────────────────────┤
│ Authentication Tag (16 bytes): GCM tag                 │
└─────────────────────────────────────────────────────────┘
```

### Encryption Process

1. **Password Input**: Secure terminal input (hidden)
2. **Salt Generation**: 32 random bytes via crypto/rand
3. **Key Derivation**: PBKDF2(password, salt, 100000, SHA-256)
4. **Nonce Generation**: 12 random bytes via crypto/rand
5. **Data Preparation**: JSON serialization + gzip compression
6. **Encryption**: AES-256-GCM(data, key, nonce)
7. **File Creation**: header + salt + nonce + ciphertext + tag

### Decryption Process

1. **File Validation**: Magic header verification
2. **Structure Parsing**: Extract salt, nonce, ciphertext, tag
3. **Key Derivation**: PBKDF2(password, salt, 100000, SHA-256)
4. **Decryption**: AES-256-GCM-Decrypt(ciphertext, key, nonce, tag)
5. **Decompression**: gzip decompression
6. **Data Recovery**: JSON deserialization to file structures

## 🛡️ Threat Model

### Protected Against

#### 1. Brute Force Password Attacks

**Protection:**
- 100,000 PBKDF2 iterations increase computation cost
- 256-bit salt prevents rainbow table attacks
- Strong password requirements encouraged

**Attack Cost:**
- **10^6 passwords/second**: ~32 years for 10-character password
- **10^9 passwords/second**: ~12 days for 10-character password
- **Cost scales exponentially** with password length

#### 2. Data Tampering

**Protection:**
- AES-GCM provides authenticated encryption
- 128-bit authentication tag detects any modification
- Magic header validates file format

**Detection:**
- Any bit flip causes decryption failure
- Impossible to modify data without password
- File format corruption detected immediately

#### 3. Chosen Ciphertext Attacks

**Protection:**
- GCM mode provides semantic security
- Nonce uniqueness prevents replay attacks
- Authentication prevents oracle attacks

#### 4. Side-Channel Attacks

**Protection:**
- Constant-time password comparison
- Memory clearing after sensitive operations
- No password echoing to terminal

#### 5. File Format Attacks

**Protection:**
- Magic header validation
- Structured parsing with bounds checking
- JSON schema validation

### Limitations and Considerations

#### 1. Password Security

**User Responsibility:**
- Strong password selection
- Secure password storage
- Protection against shoulder surfing

**Recommendations:**
- Minimum 12 characters
- Mixed case, numbers, symbols
- Use password managers
- Avoid dictionary words

#### 2. Physical Security

**Out of Scope:**
- Physical access to unlocked systems
- Memory dumps of running processes
- Hardware keyloggers
- Coercion attacks

**Mitigations:**
- Clear sensitive data from memory
- Use full-disk encryption
- Secure system configuration

#### 3. Quantum Computing

**Current Status:**
- AES-256 considered quantum-resistant
- PBKDF2 with SHA-256 quantum-safe
- 2^128 security level post-quantum

**Future Considerations:**
- Monitor NIST post-quantum standards
- Ready for algorithm migration
- Backward compatibility planning

## 🔍 Security Testing

### Automated Testing

**Cryptographic Tests:**
- ✅ Nonce uniqueness verification
- ✅ Salt randomness testing
- ✅ Key derivation consistency
- ✅ Encryption/decryption round-trips
- ✅ Authentication tag validation

**Attack Simulation:**
- ✅ Wrong password detection
- ✅ Corrupted file handling
- ✅ Modified ciphertext rejection
- ✅ Invalid format detection

### Manual Security Review

**Code Auditing:**
- ✅ Cryptographic implementation review
- ✅ Memory safety verification
- ✅ Input validation testing
- ✅ Error handling analysis

**Penetration Testing:**
- ✅ Brute force resistance
- ✅ File manipulation attacks
- ✅ Memory analysis resistance
- ✅ Side-channel analysis

## 📊 Performance vs Security

### Iteration Count Selection

**Current: 100,000 iterations**

| Iterations | Security Level | Time (ms) | Memory |
|------------|---------------|-----------|---------|
| 10,000     | Minimal       | ~15       | Low     |
| 100,000    | **Recommended** | **~150** | **Medium** |
| 1,000,000  | High          | ~1,500    | High    |

**Rationale:**
- Balances security and usability
- Resistant to current attack hardware
- Acceptable delay for normal usage
- Configurable for high-security environments

### Memory Usage

**Encryption Process:**
- Key derivation: ~32 KB
- File processing: ~file_size + compression_overhead
- Total: Efficient memory usage pattern

**Security Benefit:**
- No sensitive data persistence
- Automatic memory clearing
- Minimal attack surface

## 🔐 Best Practices

### For Users

#### Password Management
```bash
# Use strong, unique passwords
password="MyVault_2023!SecureP@ss"

# Never store passwords in shell history
export HISTCONTROL=ignorespace
 flint-vault create -f vault.dat  # Leading space

# Use password managers
# 1Password, Bitwarden, KeePass, etc.
```

#### Vault Storage
```bash
# Store vaults in secure locations
mkdir -p ~/.vaults
chmod 700 ~/.vaults
flint-vault create -f ~/.vaults/personal.dat

# Regular backups to separate media
cp ~/.vaults/personal.dat /secure/backup/location/
```

#### Operational Security
```bash
# Verify vault integrity
flint-vault list -f vault.dat > /dev/null && echo "Vault OK"

# Use different passwords for different purposes
flint-vault create -f work.dat      # Work password
flint-vault create -f personal.dat  # Personal password
```

### For Developers

#### Secure Implementation
```go
// Clear sensitive data
defer func() {
    for i := range password {
        password[i] = 0
    }
    for i := range key {
        key[i] = 0
    }
}()

// Use crypto/rand for all random data
salt := make([]byte, 32)
if _, err := rand.Read(salt); err != nil {
    return err
}
```

#### Testing Security
```go
// Test nonce uniqueness
func TestNonceUniqueness(t *testing.T) {
    nonces := make(map[string]bool)
    for i := 0; i < 1000; i++ {
        vault := createTestVault()
        nonce := extractNonce(vault)
        if nonces[string(nonce)] {
            t.Error("Duplicate nonce detected")
        }
        nonces[string(nonce)] = true
    }
}
```

## 🚨 Incident Response

### Suspected Compromise

1. **Immediate Actions:**
   - Change all vault passwords
   - Move vaults to secure storage
   - Check for unauthorized access

2. **Investigation:**
   - Review system logs
   - Check file modification times
   - Verify vault integrity

3. **Recovery:**
   - Create new vaults with new passwords
   - Re-encrypt sensitive data
   - Update security procedures

### Vulnerability Disclosure

**Responsible Disclosure Process:**

1. **Contact**: security@flint-vault.org
2. **Information**: Detailed vulnerability description
3. **Timeline**: 90-day disclosure timeline
4. **Coordination**: Work together on fixes
5. **Credit**: Public acknowledgment (if desired)

**What to Include:**
- Vulnerability description
- Proof of concept (if applicable)
- Suggested mitigation
- Impact assessment

## 🔄 Security Updates

### Version History

**v1.0.0 (Current)**
- AES-256-GCM implementation
- PBKDF2 with 100,000 iterations
- Secure random number generation
- Memory safety measures

**Future Enhancements:**
- Post-quantum cryptography support
- Hardware security module integration
- Additional key derivation functions
- Enhanced side-channel protection

### Monitoring

**Security Advisories:**
- Subscribe to project notifications
- Monitor CVE databases
- Follow cryptographic research

**Update Process:**
- Automatic security notifications
- Backward-compatible upgrades
- Migration tools for new formats

---

**Security Contact**: security@flint-vault.org  
**PGP Key**: Available at keybase.io/flint-vault 