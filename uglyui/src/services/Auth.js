import { UserManager, WebStorageStateStore } from "oidc-client";
import store from "@/store";
import Auth0 from "./Auth-auth0";
import Cognito from "./Auth-cognito";
import { ConcurrentActionHandler } from "@/utils/ConcurrentActionHandler.js";

const authCanRefreshKey = "authCanRefreshKey";

function authClient() {
  const loginRequiredError = new Error("Login Required");

  let provider;
  switch (process.env.VUE_APP_AUTH_PROVIDER) {
    case "auth0":
      provider = Auth0;
      break;
    case "cognito":
      provider = Cognito;
      break;
    default:
      throw new Error("Invalid process.env.VUE_APP_AUTH_PROVIDER: " + process.env.VUE_APP_AUTH_PROVIDER);
  }

  const mgr = new UserManager({
    userStore: new WebStorageStateStore({ store: window.sessionStorage }),
    ...provider.config()
  });

  const concurrentActionHandler = new ConcurrentActionHandler();

  mgr.getUser().then(user => {
    store.dispatch("userSet", user ? provider.standardizeUser(user) : null);
  });

  function login() {
    return mgr.signinRedirect();
  }

  function logout() {
    sessionStorage.removeItem(authCanRefreshKey);
    return mgr.signoutRedirect();
  }

  async function getAccessToken() {
    const user = await mgr.getUser();
    if (user && user.access_token) {
      return user.access_token;
    }
    return refreshAccessToken();
  }

  // for testing; make the token invalid
  async function expireAccessToken() {
    const user = await mgr.getUser();
    if (user) {
      user.access_token = "x" + user.access_token + "x";
      mgr.storeUser(user);
    }
  }

  // for testing; make the token invalid
  async function expireRefreshToken() {
    const user = await mgr.getUser();
    if (user) {
      user.refresh_token = "x" + user.refresh_token + "x";
      mgr.storeUser(user);
    }
  }

  async function refreshAccessToken() {
    let user = await mgr.getUser();
    if ((user && user.refresh_token) || (provider.canSilentlyRefresh() && sessionStorage.getItem(authCanRefreshKey))) {
      // Refresh the access token
      // The concurrency handler will only do the refresh work for the first UI view that requests it
      await concurrentActionHandler.execute(performTokenRefresh);
      user = await mgr.getUser();
      if (user && user.access_token) {
        return user.access_token;
      }
    }
    await mgr.removeUser();
    sessionStorage.removeItem(authCanRefreshKey);
    store.dispatch("userSet", null);
    throw loginRequiredError;
  }

  async function performTokenRefresh() {
    try {
      // Call the OIDC Client method
      await mgr.signinSilent();
    } catch (e) {
      // clear token data and return success, to force a login redirect
      await mgr.removeUser();
      sessionStorage.removeItem(authCanRefreshKey);
      store.dispatch("userSet", null);
    }
  }

  return {
    loginRequiredError,
    login,
    logout,
    getAccessToken,
    refreshAccessToken,
    expireAccessToken,
    expireRefreshToken
  };
}

export default authClient();
