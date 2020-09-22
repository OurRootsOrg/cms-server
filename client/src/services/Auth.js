import { UserManager, WebStorageStateStore } from "oidc-client";
import Server from "@/services/Server.js";
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

  const storageManager = window.localStorage;
  const mgr = new UserManager({
    userStore: new WebStorageStateStore({ store: storageManager }),
    ...provider.config()
  });

  const concurrentActionHandler = new ConcurrentActionHandler();

  let loading = new Promise(resolve => {
    mgr.getUser().then(user => {
      console.log("getUser", user);
      // store.dispatch("userSet", user ? provider.standardizeUser(user) : null);
      // resolve(true);
      // loading = null;
      if (user) {
        // verify token has not expired
        Server.currentUser().then(
          response => {
            let stdUser = Object.assign({}, response.data, provider.standardizeUser(user));
            console.log("stdUser", stdUser);
            store.dispatch("userSet", stdUser);
            console.log("currentUser set");
            resolve(true);
            loading = null;
          },
          err => {
            console.log("currentUser error", err);
            mgr.removeUser().then(() => {
              console.log("user removed");
            });
            storageManager.removeItem(authCanRefreshKey);
            store.dispatch("userSet", null);
            resolve(true);
            loading = null;
          }
        );
      } else {
        storageManager.removeItem(authCanRefreshKey);
        store.dispatch("userSet", null);
        resolve(true);
        loading = null;
      }
    });
  });

  async function isLoaded() {
    if (loading) {
      return loading;
    } else {
      return true;
    }
  }

  function login() {
    return mgr.signinRedirect();
  }

  function logout() {
    storageManager.removeItem(authCanRefreshKey);
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
    console.log("refreshAccessToken", user);
    if ((user && user.refresh_token) || (provider.canSilentlyRefresh() && storageManager.getItem(authCanRefreshKey))) {
      // Refresh the access token
      // The concurrency handler will only do the refresh work for the first UI view that requests it
      await concurrentActionHandler.execute(performTokenRefresh);
      user = await mgr.getUser();
      console.log("refreshAccessToken refreshed", user);
      if (user && user.access_token) {
        return user.access_token;
      }
    }
    await mgr.removeUser();
    storageManager.removeItem(authCanRefreshKey);
    store.dispatch("userSet", null);
    throw loginRequiredError;
  }

  async function performTokenRefresh() {
    try {
      // Call the OIDC Client method
      console.log("performTokenRefresh");
      await mgr.signinSilent();
      console.log("performTokenRefresh success");
    } catch (e) {
      // clear token data and return success, to force a login redirect
      console.log("performTokenRefresh error", e);
      await mgr.removeUser();
      storageManager.removeItem(authCanRefreshKey);
      store.dispatch("userSet", null);
    }
  }

  return {
    loginRequiredError,
    isLoaded,
    login,
    logout,
    getAccessToken,
    refreshAccessToken,
    expireAccessToken,
    expireRefreshToken
  };
}

export default authClient();
