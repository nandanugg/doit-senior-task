# ADR 001: Hex Code Masking for ID Obfuscation

**Status:** Accepted

**Date:** 2026-01-11

**Context:** modules/core/lib/hex.go

## Context

When designing URL-safe identifiers for our API, we need to balance several concerns: database performance, CPU efficiency, and ID obfuscation. This ADR documents our decision to use hex encoding with a simple character replacement map.

### Storage Strategy

We store URL mappings in **Redis** using an atomic `INCR` counter for IDs. This keeps ID allocation fast and consistent across replicas while avoiding index maintenance costs entirely. The same sequential IDs are stored in PostgreSQL for analytics, which benefits from predictable ordering.

## Decision

We chose to use **hex encoding with character replacement** to obfuscate sequential IDs in URLs and API responses.

### Why Not Base64?

Base64 and its variants (base62, base58) are commonly used for ID encoding:
- **Base64**: Encodes 6 bits per character, very efficient
- **CPU cost**: Requires bit shifting and lookup tables (typically ~10-15 CPU cycles per character)
- **URL-safe variants**: base64url replaces `+` and `/` with `-` and `_`

While base64 is proven and efficient, we wanted to **experiment with a simpler approach** for this codebase.

### Hex Encoding Approach

We use hexadecimal encoding with selective character replacement:

1. **Convert ID to hex**: `fmt.Sprintf("%x", n)` - native Go operation, highly optimized (~5 CPU cycles)
2. **Apply replacement map**: Replace ambiguous characters to improve readability
   ```go
   '0' → 'g'  // Avoid confusion with letter 'O'
   '1' → 'h'  // Avoid confusion with letter 'I' or 'l'
   ```
3. **CPU cost**: Simple rune iteration and map lookup (~3-5 CPU cycles per character)

**Total CPU cost**: Approximately **8-10 CPU cycles per character** - comparable to base64 but with simpler implementation.

### Character Ambiguity Handling

The replacement map addresses common visual ambiguities in URL-safe identifiers:
- `0` (zero) vs `O` (letter O)
- `1` (one) vs `I` (letter i) vs `l` (lowercase L)

The current map is minimal but can be extended:
```go
var encodeMap = map[rune]rune{
    '0': 'g',
    '1': 'h',
    // Additional mappings can be added:
    // 'o': 'q',
    // 'i': 'j',
}
```

## Consequences

### Positive

- **Simple implementation**: Easy to understand and maintain
- **Efficient**: Near-native hex encoding performance
- **Readable**: Avoids common character confusion in URLs
- **Debuggable**: Hex patterns are familiar to developers

### Negative

- **Security concern**: Simple character substitution is trivially reversible
- **Weak obfuscation**: The replacement map is easily discovered through:
  - Analyzing multiple IDs to identify patterns
  - Simple frequency analysis
  - Brute-force mapping (only 16 hex characters)
- **Length**: Hex encoding is less compact than base64 (4 bits vs 6 bits per character)

### Security Considerations

**When this approach is acceptable:**
- Global consumer services (like bit.ly) where obfuscation prevents casual enumeration
- Systems where ID exposure is low-risk
- Applications with proper authorization checks at the API level

**When this approach is NOT advisable:**
- **Enterprise applications** with sensitive data
- Systems requiring cryptographic security
- Applications where ID predictability poses business risk
- Multi-tenant systems where tenant isolation depends on ID secrecy

**Important**: This implementation provides **obfuscation, not security**. A simple hex-to-decimal conversion followed by character mapping reversal will expose sequential IDs. Our security model must not rely on ID secrecy - always implement proper authorization checks.

### Growing the Replacement Table

**Evolution Path**: As long as we maintain **numeric IDs in the database**, we can evolve this encoding scheme to higher bases with larger character maps:

- **Current**: Base16 (hex) with 16-character alphabet
  - Simple, readable, debuggable
  - ~4 bits per character

- **Future options**:
  - **Base32**: 32-character alphabet (~5 bits per character) - more compact, still readable
  - **Base58**: 58-character alphabet (Bitcoin-style) - excludes ambiguous chars (0, O, I, l)
  - **Base64**: 64-character alphabet (~6 bits per character) - maximum compactness
  - **Custom bases**: Any alphabet size with custom replacement maps

**Database independence**: The key architectural decision is keeping **integer IDs at the database level**. This means:
- Database queries use standard integer operations (fast, well-indexed)
- Encoding/decoding happens at the application layer
- We can change the encoding scheme without database migrations
- Different API versions can use different encoding schemes simultaneously

**Security caveat**: While extending to larger character maps, remember this remains **simple substitution cipher**:
- A 16-character hex map has at most 16! permutations, but even a small sample of IDs reveals the pattern
- Base64 with a 64-character map is equally vulnerable to frequency analysis
- Adding complexity hurts readability without adding real security
- **If security matters, use cryptographic approaches** (UUIDs, encrypted IDs, or signed tokens)

The replacement table can grow, but don't mistake increased complexity for increased security. For enterprise applications with sensitive data, proper cryptographic protection is the only viable option.

## Alternatives Considered

1. **UUID v4**: Cryptographically random, but wastes database index space and is harder to debug
2. **Base64 encoding**: More compact, but no significant advantage given our use case
3. **Hashids library**: More sophisticated obfuscation, but still reversible and adds dependency
4. **Encrypted IDs**: Proper security, but overkill for our current threat model

## References

- Implementation: `modules/core/lib/hex.go:11-59`
- Base64 performance comparison: https://lemire.me/blog/2018/01/17/ridiculously-fast-base64-encoding-and-decoding/
