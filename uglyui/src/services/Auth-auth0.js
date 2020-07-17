export default {
  config() {
    return {
      authority: process.env.VUE_APP_AUTH_DOMAIN,
      client_id: process.env.VUE_APP_AUTH_CLIENT_ID,
      redirect_uri: process.env.VUE_APP_AUTH_REDIRECT_URL,
      response_type: "id_token token",
      scope: "openid profile email",
      filterProtocolClaims: true,
      // Auth0 doesn't include an end_session_endpoint so we have to list everything
      metadata: {
        issuer: process.env.VUE_APP_AUTH_DOMAIN + "/",
        authorization_endpoint: process.env.VUE_APP_AUTH_ENDPOINT_DOMAIN + "/authorize",
        token_endpoint: process.env.VUE_APP_AUTH_ENDPOINT_DOMAIN + "/oauth/token",
        userinfo_endpoint: process.env.VUE_APP_AUTH_ENDPOINT_DOMAIN + "/userinfo",
        jwks_uri: process.env.VUE_APP_AUTH_DOMAIN + "/.well-known/jwks.json",
        end_session_endpoint:
          process.env.VUE_APP_AUTH_ENDPOINT_DOMAIN +
          "/v2/logout?returnTo=" +
          encodeURIComponent(process.env.VUE_APP_AUTH_POST_LOGOUT_REDIRECT_URL)
      }
    };
  },
  standardizeUser(user) {
    return {
      name: user.profile.nickname,
      email: user.profile.name,
      picture: user.profile.picture
    };
  },
  canSilentlyRefresh() {
    return true;
  }
};
