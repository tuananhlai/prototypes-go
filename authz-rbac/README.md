# Common Authorization Mechanisms

Authorization mechanisms define **how access to resources is decided after authentication**.  
RBAC is common, but it’s only one of several major models.

---

## 1. RBAC (Role-Based Access Control)
**Access = user → role → permissions**

**Example**
- Roles: `admin`, `editor`, `viewer`
- Permissions: `read_article`, `publish_article`

**Pros**
- Simple
- Easy to reason about
- Widely supported

**Cons**
- Role explosion in complex systems
- Hard to model context (time, location, ownership)

**Best for**
- Internal tools
- Organizations with stable roles

---

## 2. ABAC (Attribute-Based Access Control)
**Access is decided by attributes + policies**

**Attributes**
- User: department, clearance level
- Resource: classification, owner
- Environment: time, IP, location

**Policy example**
- Allow if `user.department == "finance"`
- AND `resource.type == "report"`
- AND `time < 6pm`

**Pros**
- Very flexible
- Fine-grained control

**Cons**
- Complex to design
- Harder to debug

**Best for**
- Enterprises
- Regulatory environments (finance, healthcare)

---

## 3. PBAC (Policy-Based Access Control)
**Access controlled by explicit policies (often built on ABAC)**

**Example (conceptual)**
- Effect: allow
- Action: read
- Resource: invoice
- Condition: user.role == manager

**Pros**
- Centralized rules
- Clear auditability

**Cons**
- Policy engines add complexity

**Best for**
- Cloud platforms (AWS IAM, Azure, GCP)

---

## 4. ACL (Access Control List)
**Each resource lists who can access it**

**Example**
- file.txt
  - Alice: read
  - Bob: read, write

**Pros**
- Simple and intuitive
- Fine-grained per resource

**Cons**
- Hard to manage at scale
- Duplication of rules

**Best for**
- Filesystems
- Small-scale systems

---

## 5. DAC (Discretionary Access Control)
**Resource owner decides access**

**Example**
- A document shared by its owner

**Pros**
- Flexible
- User-driven

**Cons**
- Security risks
- Hard to enforce global rules

**Best for**
- Collaboration tools

---

## 6. MAC (Mandatory Access Control)
**System-enforced security levels**

**Example**
- Military classifications: Top Secret, Secret, Confidential

**Pros**
- Very secure
- No user override

**Cons**
- Rigid
- Expensive to manage

**Best for**
- Defense
- High-security systems

---

## 7. ReBAC (Relationship-Based Access Control)
**Access based on relationships between entities**

**Example**
- User owns a document
- User is a member of a team
- User manages another user

**Policy example**
- Allow if user is owner of document
- OR user is in document team

**Pros**
- Natural modeling for modern apps
- Scales well for graph-like relationships

**Cons**
- Requires graph-oriented thinking

**Best for**
- SaaS products
- Collaboration platforms

---

## 8. Token / Scope-Based Authorization (OAuth-style)
**Access via tokens with scopes**

**Example**
- read:profile
- write:posts

**Pros**
- Stateless
- Great for APIs and microservices

**Cons**
- Limited expressiveness without policies

**Best for**
- APIs
- Mobile and third-party integrations

---

## 9. Capability-Based Access Control
**Possession of a capability grants access**

**Example**
- Signed URLs
- Temporary access tokens

**Pros**
- Simple
- Works well in distributed systems

**Cons**
- Revocation is hard
- Token leakage risk

**Best for**
- Object storage
- Temporary access links

---

## High-level Comparison

| Model | Flexibility | Complexity | Common Use |
|------|-------------|------------|------------|
| RBAC | Low–Medium | Low | Internal apps |
| ABAC | Very High | High | Enterprise |
| PBAC | High | Medium–High | Cloud IAM |
| ACL | Medium | Low | Filesystems |
| ReBAC | High | Medium | SaaS apps |
| OAuth Scopes | Medium | Low | APIs |

---

## 9. Minimal RBAC Implementation Example

This directory contains a minimal Go implementation of RBAC.

### How to Run
```bash
go run main.go
```

### How to Test
This implementation uses JWT for identity and RBAC for authorization.

**1. Login to get a JWT:**
```bash
# Get token for Alice (Admin)
curl -X POST "http://localhost:8080/login?user=alice"
# Get token for Charlie (Viewer)
curl -X POST "http://localhost:8080/login?user=charlie"
```

**2. Access protected endpoint (Success):**
Copy the token from step 1 and use it in the Authorization header.
```bash
TOKEN="your_token_here"
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/data
```

**3. Access restricted action (Forbidden):**
If you use Charlie's token to POST, it will be forbidden.
```bash
TOKEN_CHARLIE="charlie_token"
curl -X POST -H "Authorization: Bearer $TOKEN_CHARLIE" http://localhost:8080/data
```

---

## Why "Permission" is needed?
In this implementation, the JWT contains the `Role` (e.g., "viewer"). However, the code checks for a `Permission` (e.g., "read:data").

1. **Decoupling:** If you decide that `viewers` should also be able to `write`, you only change the `rolePermissions` map. You don't have to touch the logic in `handleWrite`.
2. **Evolution:** You can add a new role (e.g., "Manager") and simply assign it the existing permissions without changing any API handlers.
3. **Auditability:** It's easier to audit "who can write data" by looking at the permission mapping than by hunting through `if role == "admin" || role == "editor"` checks.
