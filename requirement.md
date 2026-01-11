# Distributed URL Shortener - Technical Requirements

## Overview

Design and implement a URL shortening service that demonstrates architectural patterns, operational safety, and infrastructure design.

**Estimated Time**: 3-5 hours focused effort

**Important**: If you reach the 5-hour mark, pause development and document remaining implementation details and design decisions in the README.

---

## 1. Functional Requirements

### 1.1 URL Creation

**Endpoint**: Accept `long_url` and optional `ttl_seconds` (default: 24 hours)

**Input Validation**:
- Validate that the provided string is a valid HTTP/HTTPS URL
- Enforce reasonable length constraints to mitigate system abuse

**Uniqueness Strategy**:
- Define a short-code generation strategy that ensures uniqueness
- Document how your implementation handles hash collisions
- Document how it maintains integrity at scale

**Character Constraint**:
- Generated short codes must **exclude** these characters for human readability and to avoid visual ambiguity:
  - `0` (zero)
  - `O` (capital letter O)
  - `I` (capital letter I)
  - `l` (lowercase letter L)
  - `1` (one)

### 1.2 Redirection Logic

**Endpoint**: `GET /s/{short_code}`

**Behavior**:
- Execute a **302 redirect** for valid, active codes
- Return **404 status** for nonexistent or expired records

**Persistence**:
- Update `click_count` and `last_accessed_at` in a thread-safe manner

### 1.3 Observability and Metrics

**Endpoint**: `GET /stats/{short_code}`

**Response**: Return full record containing:
- `long_url`
- `created_at`
- `expires_at`
- `click_count`
- `last_accessed_at`

**Custom Instrumentation**:
- All responses (Success, Redirect, or Error) must include custom HTTP header:
  - `X-Processing-Time-Micros`: Internal execution duration in microseconds

**Data Privacy**:
- System must adhere to privacy-by-design principles
- **PII (Personally Identifiable Information)**, specifically requester IP addresses, must **not** be:
  - Persisted in the storage layer
  - Exposed in system logs

### 1.4 Expiration Management

**Implementation**:
- Implement a strategy for handling expired records
- Choose between: Lazy, Background, or Hybrid cleanup mechanism

**Documentation**:
- Document your chosen cleanup mechanism
- Justify the associated operational tradeoffs

---

## 2. Technical Requirements

### 2.1 Software Architecture

**Storage Implementation**:
- You may use an in-memory storage implementation for this exercise
- Architecture must demonstrate strict separation of concerns
- Use an abstraction layer (Interface or Abstract Class)
- Storage engine must be transitionable to production database (DynamoDB, Firestore, Redis) without modifying business logic

**Concurrency**:
- Employ idiomatic concurrency primitives
- Manage shared state and prevent data races during high-frequency updates

### 2.2 Static Analysis and Quality Standards

**CI/CD Pipeline Required** (GitHub Actions):
- Pipeline must enforce strict quality gates
- Pipeline must **fail** if any warnings are detected

**Quality Gates**:
- **Complexity**: No single function may exceed Cyclomatic Complexity of 10

**Language-Specific Requirements**:

**Go**:
- Must pass `golangci-lint` with `gocyclo` and `revive` linters enabled

**Python**:
- Must pass `mypy --strict`
- Must pass `ruff`

**TypeScript**:
- Must pass `eslint` (with `no-explicit-any`)
- Must pass `tsconfig` with `strict: true`

### 2.3 Containerization and Infrastructure

**Docker**:
- Provide a multi-stage Dockerfile
- For security compliance, container must run as a **non-root user**

**Infrastructure as Code**:
- Provide a syntactically valid Terraform (`main.tf`) file
- Define production environment on either **AWS** or **GCP**
- Must include:
  - Serverless compute resource
  - Managed storage resource
  - Associated IAM roles following the **Principle of Least Privilege**

### 2.4 Testing Standards (Signal over Coverage)

**Note**: Submissions are not evaluated based on line coverage percentages. Focus on high-signal testing.

**Required Tests**:

1. **Concurrency Validation**:
   - Include a test demonstrating `click_count` logic is safe under 100+ concurrent requests

2. **Deterministic Expiration**:
   - Verify TTL expiration logic by mocking the system clock
   - Test suite must be deterministic
   - Must **not** utilize `sleep` or real-time delays

3. **Interface Verification**:
   - Demonstrate that domain logic is tested against the storage abstraction (Interface/Abstract Class)
   - Do not test against concrete implementation

---

## 3. Evaluation Criteria

Your submission will be evaluated on:

- **Engineering Judgment**: Balance speed of delivery with correctness and long-term maintainability
- **Operational Maturity**: Evidence of proactive thinking regarding observability, capacity planning, and data privacy
- **Testing Discipline**: Prioritization of high-signal tests (concurrency, deterministic time) over superficial line coverage
- **Tooling Proficiency**: Competency in Git workflows, Docker configuration, and Infrastructure as Code
- **Decomposition**: Ability to translate ambiguous requirements into modular, testable code

---

## 4. Submission Instructions

### Git History Requirement

**Critical**: Your submission must be a Git repository (or compressed archive containing `.git` directory) with clear, incremental commit history.

**⚠️ Disqualification**: Submissions containing a single "Initial Commit" will be disqualified.

### Documentation (README)

Your README must include:

#### 1. Architecture Overview
- Description of your data model
- High-level system diagram

#### 2. Gap Analysis
- Technical explanation of why in-memory implementation is unsuitable for stateless cloud environment defined in Terraform
- How your proposed managed storage solves this

#### 3. Capacity Planning
- Technical estimate for storage requirements (in GB) over 12-month period
- Assumption: 100 million new URLs per month
- Scaling strategy for redirect endpoint to handle 10,000 requests per second

#### 4. Service Level Management
- Define two Service Level Indicators (SLIs) for this service
- Propose a Service Level Objective (SLO) for each
- Describe one specific scenario that would necessitate on-call intervention

#### 5. Future Enhancements
- Summary of features or architectural improvements you would prioritize with additional time
