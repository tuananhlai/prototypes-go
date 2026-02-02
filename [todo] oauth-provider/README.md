An OAuth provider is a service that issues access tokens to apps after a user authorizes them.
Those tokens let the app call APIs without ever seeing the user’s password.

Let’s walk it end-to-end, clean and practical.

⸻

The roles (4 actors) 1. User – the human 2. Client – your app (web / mobile / backend) 3. OAuth Provider – Google, GitHub, Auth0, etc. 4. Resource Server – API that holds the data
(often the same system as the provider)

⸻

What the provider actually does

An OAuth provider:
• Authenticates the user
• Shows a consent screen
• Issues tokens (access + optional refresh)
• Validates and introspects tokens
• Enforces scopes & expiry

⸻

The most common flow: Authorization Code (with PKCE)

This is what browsers, SPAs, and mobile apps use.

⸻

1️⃣ App redirects user to provider

Your app sends the user to the provider’s authorization endpoint:

GET /authorize?
response_type=code
&client_id=abc123
&redirect_uri=https://app.com/callback
&scope=profile email
&state=xyz
&code_challenge=...
&code_challenge_method=S256

Nothing secret here.

⸻

2️⃣ Provider authenticates the user

The provider:
• Shows login (password, SSO, MFA, passkeys, etc.)
• Identifies the user

⸻

3️⃣ Provider asks for consent

User sees:

“App X wants access to your profile and email”

User clicks Allow or Deny.

⸻

4️⃣ Provider redirects back with a code

If approved:

https://app.com/callback?
code=AUTH_CODE_123
&state=xyz

The code:
• Is short-lived
• Can be used only once
• Is useless without PKCE / client secret

⸻

5️⃣ App exchanges code for tokens (backend call)

Your app calls the token endpoint:

POST /token
grant_type=authorization_code
code=AUTH_CODE_123
client_id=abc123
code_verifier=...

⸻

6️⃣ Provider validates and issues tokens

Provider checks:
• Code validity
• PKCE proof
• Client identity
• Redirect URI

Then returns:

{
"access_token": "eyJhbGciOi...",
"token_type": "Bearer",
"expires_in": 3600,
"refresh_token": "def456",
"scope": "profile email"
}

⸻

7️⃣ App uses access token

App calls APIs:

Authorization: Bearer <access_token>

The resource server:
• Verifies signature / introspects token
• Checks scopes
• Returns data

⸻

What the OAuth provider maintains internally (state)
• Client registrations (client_id, redirect URIs)
• User identities
• Granted consents
• Issued tokens & expiry
• Refresh token rotation
• Revocations

⸻

Access token vs Refresh token

Token Purpose
Access token Short-lived API access
Refresh token Get new access tokens without user login

Access tokens expire fast → safer
Refresh tokens are tightly protected

⸻

OAuth ≠ Authentication (important)

OAuth answers:

“Is this app allowed to access this resource?”

Not:

“Who is the user?”

That’s why OpenID Connect (OIDC) exists.

⸻

OAuth + OIDC (what most people actually use)

OIDC adds:
• id_token (JWT with user identity)
• Standard claims (sub, email, name)
• /userinfo endpoint

This is what “Login with Google” really is.

⸻

Why OAuth providers exist
• Users don’t share passwords
• Apps get least-privilege access
• Tokens can be revoked
• Centralized security (MFA, risk checks)

⸻

TL;DR

An OAuth provider: 1. Authenticates users 2. Gets user consent 3. Issues short-lived access tokens 4. Enforces scopes & expiry 5. Lets apps access APIs safely
