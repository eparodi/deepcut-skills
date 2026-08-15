# Security Knowledge Base (reference)

> Extracted from the `security-engineer` skill (2026-08-15) — generic
> knowledge every model already has. Load ON DEMAND, never bundled
> with the skill. Keep in sync when the skill's checklists change.

## Security Knowledge Base

### OWASP Top 10 (2021)

| # | Category | Description |
|---|----------|-------------|
| A01 | Broken Access Control | Missing or improper authorization checks, IDOR, path traversal |
| A02 | Cryptographic Failures | Weak encryption, hardcoded keys, missing TLS, weak randomness |
| A03 | Injection | SQLi, NoSQLi, command injection, LDAP injection |
| A04 | Insecure Design | Missing rate limiting, no threat modeling, missing security controls |
| A05 | Security Misconfiguration | Default credentials, verbose errors, missing security headers, open S3 buckets |
| A06 | Vulnerable Components | Outdated dependencies with known CVEs |
| A07 | Auth Failures | Weak passwords, missing MFA, session fixation, credential stuffing |
| A08 | Software & Data Integrity | Unsafe deserialization, CI/CD pipeline attacks, unsigned dependencies |
| A09 | Logging & Monitoring Failures | No audit logs, no alerting, log injection |
| A10 | SSRF | Server-Side Request Forgery via user-controlled URLs |

### OWASP API Security Top 10 (2023)

| # | Category |
|---|----------|
| API1 | Broken Object Level Authorization (BOLA/IDOR) |
| API2 | Broken Authentication |
| API3 | Broken Object Property Level Authorization (Mass Assignment) |
| API4 | Unrestricted Resource Consumption (Rate Limiting) |
| API5 | Broken Function Level Authorization |
| API6 | Unrestricted Access to Sensitive Business Flows |
| API7 | Server-Side Request Forgery (SSRF) |
| API8 | Security Misconfiguration |
| API9 | Improper Inventory Management |
| API10 | Unsafe Consumption of APIs |

### Common Attack Vectors

- **XSS (Cross-Site Scripting)**: Reflected, Stored, DOM-based
- **CSRF (Cross-Site Request Forgery)**: State-changing requests without anti-CSRF tokens
- **IDOR (Insecure Direct Object Reference)**: Accessing resources by predictable IDs without auth checks
- **Path Traversal**: `../../../etc/passwd` in file paths
- **Command Injection**: User input passed to `exec`, `system`, `os/exec`
- **JWT Attacks**: None algorithm, weak HMAC secrets, expired token handling, key confusion
- **WebSocket Attacks**: No auth on WS upgrade, origin spoofing, message forgery
- **Timing Attacks**: Constant-time comparison failures
- **Race Conditions**: TOCTOU (Time-of-check to time-of-use), double-spend
- **DoS/DDoS**: Unbounded resource allocation, missing timeouts, missing body size limits

### Security Headers Checklist

```
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
Content-Security-Policy: default-src 'self'
X-XSS-Protection: 0  (deprecated, use CSP instead)
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: camera=(), microphone=(), geolocation=()
```

### Cryptographic Standards

- **Passwords**: bcrypt (cost >= 12), argon2id, scrypt. NEVER MD5, SHA1, SHA256 (unsalted).
- **Tokens/Keys**: crypto/rand (Go), crypto.randomBytes (Node) — never `math/rand`.
- **JWT**: ES256 or RS256 preferred. HMAC only if the secret is >= 256 bits and not shared. Reject `alg: none`.
- **TLS**: Minimum TLS 1.2, prefer TLS 1.3. No weak ciphers.
- **Hash Functions**: SHA-256 (minimum), SHA-384, SHA-512. No MD5, SHA-1.
- **Long-Lived Secret Storage** (API keys, stream keys): Hash before storing (SHA-256), never store raw keys.

### Go-Specific Security Patterns

- Always use `crypto/rand` not `math/rand` for secrets.
- SQL: parameterized queries (`$1`, `$2`), never `fmt.Sprintf` with user input.
- Command execution: prefer `exec.Command` with args array, never shell interpolation.
- Template rendering: `html/template` (auto-escapes), not `text/template`.
- File paths: use `filepath.Clean` and validate against a base directory.
- HTTP: always set `ReadTimeout`, `WriteTimeout`, `ReadHeaderTimeout`, `IdleTimeout`.
- HTTP: always use `http.MaxBytesReader` to limit body sizes.

### TypeScript/Next.js Security Patterns

- Server Components: never expose secrets to the client (use `server-only`).
- API routes: validate all input, use Zod or similar.
- CSRF: Next.js has built-in CSRF for Server Actions, but custom API routes need manual CSRF.
- Cookie flags: `HttpOnly`, `Secure`, `SameSite=Lax` (or `Strict`).
- CSP: configure via `next.config.ts` headers or middleware.
- Image optimization: use `next/image` (prevents open redirect via image proxy).

---

