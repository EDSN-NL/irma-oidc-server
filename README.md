# IRMA-oidc-server

*Note: This project is in an early state and isn't production-ready yet, use at your own risk!*

This project is based on [fosite-example](https://github.com/ory/fosite-example). ORY Fosite is the security first OAuth2 & OpenID Connect framework for Go. 

[IRMA](https://privacybydesign.foundation/irma-en/) is a unique privacy-friendly identity platform for both authentication and signing. 

This repository combines IRMA (using [irmago](https://github.com/privacybydesign/irmago)) with OpenID Connect to provide an OpenID Connect server that authenticates users using IRMA disclosure proofs.

## Run it

Download a release from the releases page for your platform, and unpack and run it.

Visit [http://localhost:3846/oauth2/auth?scope=openid+pbdf.pbdf.email.email&response_type=code&client_id=default-client&redirect_uri=http%3A%2F%2Flocalhost%3A8080&state=veryfoobar](http://localhost:3846/oauth2/auth?scope=openid+pbdf.pbdf.email.email&response_type=code&client_id=default-client&redirect_uri=http%3A%2F%2Flocalhost%3A8080&state=veryfoobar). You should see an IRMA QR code.

See the configuration section on how to configure this for your client.

## Compile and run

With a recent version of Golang (i.e. 1.13+), dependencies are installed automatically. We only need the [packr](https://github.com/gobuffalo/packr) tool for building.

```
$ go get -u github.com/gobuffalo/packr/packr
$ packr build -o irma-oidc-server .
$ ./irma-oidc-server
```

## Configuration

Configuration of this server is done via the `config.yaml` file, that must be in the same directory as the binary. See this repository for an example file with the default values.

### Config.yaml

The following options are supported:

- irmaUrl: URL of this server, must be reachable by both phone and OpenID Connect client. Make sure it ends with '/irma'.
- port: The port to listen at
- irmaProductionMode: Set to true to enable IRMA production mode
- oidcClients: List of allowed OpenID Connect client applications:
    - id: client ID of this client
    - secret: client Secret for this secret
    - redirectUris: list of Uris a client is allowed to redirect to 
    - scopes: List of scopes: scope parameters are IRMA attribute identifiers which this client is allowed to request. Please also include the 'openid' scope to allow clients to get an `id_token` back with attribute values.
- oidcIssuer: This string will be used as issuer field for the JWT tokens.
- oauthHmacSecret: Secret used to sign OAuth authorization codes.
- jwtPrivateKey: Private key used for generating JWT tokens, will be generated if omitted.

### Example client application configuration

At this moment, only the [Authorization Code Flow](https://openid.net/specs/openid-connect-core-1_0.html#CodeFlowAuth) is supported, where an [id_token](https://openid.net/specs/openid-connect-core-1_0.html#CodeIDToken) will be returned with the IRMA attribute values. (So no `userinfo` support yet...). All attribute values will be added to the `id_token` in the `disclosed` claim. The first requested attribute value will also be added to the `subject` field of the token (so make sure this attribute uniquely identifies your users).

To configure an application to use this server, add an OpenID Connect identity provider, and use the following values:
- Authorization URL: `https://SERVER/oauth2/auth`
- Token URL: `https://SERVER/oauth2/token`
- Scopes: Add at least 'openid' and one IRMA attribute identitfier (for instance: `pbdf.pbdf.email.email`). The order of the scopes is important, because the attribute identifier first scope will 'act' as subject value in the generated `id_token`.

Generate a clientID, clientSecret and redirectUri and add them to both `config.yaml` and your client application configuration.

Replace SERVER with the public reachable url of IRMA-oidc-server.

### Example client config: Keycloak

[Keycloak](https://www.keycloak.org/) is an Open Source Identity provider, which can act as an OpenID Connect client and [federate](https://www.keycloak.org/docs/latest/server_admin/index.html#_identity_broker) authentication to another ID server (in this case the IRMA-oidc-server). 

Provided that you have a running Keycloak instance and created a new realm, take the following steps to connect this realm to IRMA-oidc-server:
- Go to Identity Providers -> Add provider (on the right, dropdown menu) -> OpenID Connect v1.0
- Fill in the Authorization URL: `https://SERVER/oauth2/auth` and Token URL: `https://SERVER/oauth2/token`
- Fill in clientID and clientSecret (generate yourself)
- Set Disable User Info to 'On'
- At the Issuer field you can fill in the `oidcIssuer` value from `config.yaml`
- Default Scopes: Add a list of IRMA attributes you want to retrieve. Make sure to place an uniquely identifiable attribute in front (since this will be mapped to the subject field / users's primary key).
    - Recommended would be: `pbdf.pbdf.email.email pbdf.gemeente.personalData.initials pbdf.gemeente.personalData.surname`, because these data can be mapped to the default user details in Keycloak.
- Validate Signatures can be set to true if you want to validate the signature on the tokens, make sure you paste the correct public key (corresponding to the `jwtPrivateKey` in `config.yaml`).

After saving, Keycloak will generate a Redirect URI for you, which you need to add to your OidcClient in `config.yaml`.

Now it's time to test: Visit https://YOUR-KEYCLOAK-SERVER/auth/realms/YOUR-REALM/account/ and press the login button, which will redirect you to the IRMA-oidc-server. You should at least be able to do a login flow this way.

However, attribute values aren't yet mapped to a Keycloak user. For this, go to the Identity Provider settings in Keycloak and visit the 'Mappers' tab. 
- Press the 'create' button to create a new mapper.
    - Fill in any name you like
    - Mapper type: Attribute Importer
    - Claim will be: `disclosed.pbdf\.pbdf\.email\.email`
    - User Attribute Name: email
- Repeat this process for the `firstName` attribute (`disclosed.pbdf\.gemeente\.personalData\.initials`) and the `lastName` attribute (`disclosed.pbdf\.gemeente\.personalData\.surname`)
- You can also add other IRMA attributes to users, map them to Keycloak roles, and so on.
