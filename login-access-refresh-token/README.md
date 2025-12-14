# Login flow with both Access and Refresh Token

- Using session cookies for login requires the server to hit the database every time it receives a session cookie, which is not ideal. We can optimize the flow by using a cache though.
- Using long-lived access tokens skip the need for a database/cache check, but you can not control them once they're issued, so revoking old access tokens is difficult when it's leaked or blacklisted. 
- For the best of both worlds (but with increased complexity), we can use a login flow with two types of token:
  1. Access token: short-lived, grant access to resources.
  2. Refresh token: long-lived, allow generation of access tokens. Managed within databases.
This way, even if an access token got leaked, the impact is minimal since it's short lived. For the refresh tokens, they can be revoked easily due to being managed within a database. While we still need to hit the database/cache once in a while to get a new access token, the frequency would be much lower compared to approach (1).
