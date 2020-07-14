export default {
  config() {
    return {
      authority: process.env.VUE_APP_AUTH_DOMAIN,
      client_id: process.env.VUE_APP_AUTH_CLIENT_ID,
      redirect_uri: process.env.VUE_APP_AUTH_REDIRECT_URL,
      response_type: "code",
      scope: "openid profile email",
      // We are not using background silent token renewal because cognito doesn't support prompt=none
      automaticSilentRenew: false,
      // We are not using these features; we get extended user info from our API
      loadUserInfo: false,
      monitorSession: false,
      // Cognito doesn't include an end_session_endpoint so we have to list everything
      metadata: {
        issuer: process.env.VUE_APP_AUTH_DOMAIN,
        authorization_endpoint: process.env.VUE_APP_AUTH_ENDPOINT_DOMAIN + "/oauth2/authorize",
        token_endpoint: process.env.VUE_APP_AUTH_ENDPOINT_DOMAIN + "/oauth2/token",
        userinfo_endpoint: process.env.VUE_APP_AUTH_ENDPOINT_DOMAIN + "/oauth2/userinfo",
        jwks_uri: process.env.VUE_APP_AUTH_DOMAIN + "/.well-known/jwks.json",
        end_session_endpoint:
          process.env.VUE_APP_AUTH_ENDPOINT_DOMAIN +
          "/logout?client_id=" +
          encodeURIComponent(process.env.VUE_APP_AUTH_CLIENT_ID) +
          "&logout_uri=" +
          encodeURIComponent(process.env.VUE_APP_AUTH_POST_LOGOUT_REDIRECT_URL)
      }
    };
  },
  standardizeUser(user) {
    return {
      name: user.profile.name,
      email: user.profile.email
    };
  },
  canSilentlyRefresh() {
    return false;
  }
};
