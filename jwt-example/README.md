# Notes
- A JWT is comprised of three parts:
    - Header: Type of token and the algorithm used
    - Payload: contains application specific data (username, user id, etc)
    - Signature: encoded header, encoded payload, and secret you provide are used to create the signature
- Token types
    - Access token: used for requests that require authentication, generally a short lifespan (15 mins)
    - Refresh token: has a longer lifespan (usually 7 days). Used when the access token expires to create new access and refresh tokens

## JWT: Things to remember
- If someone were to steal a JWT, you are out of luck, because of the statelessness of them. To circumvent that, you can add some state to the system or use refresh tokens
- Refresh token: A JWT token which you store in a database. Usually long lived, might not even have an expiry time. This has to be stored very securely if in the browser
- Access token: A JWT token that is generally short lived 

## What I am learning from this
- With access tokens you should store something unique to the user in the JWT payload. Not only should you check if the access token is still valid, but you should also check if the user issuing the request has the right to do that
- State of refresh token in some persistence layer should attribute to if the user is "logged in" or not
- We should make refreshing of access and refresh tokens opaque to the user, upon only authenticated requests
