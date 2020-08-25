import Auth from "@/services/Auth";
import axios from "axios";
import axiosRetry from "axios-retry";
const axiosClient = axios.create({
  baseURL: process.env.VUE_APP_API_BASE_URL,
  withCredentials: false, // This is the default
  headers: {
    Accept: "application/json",
    "Content-Type": "application/json"
  }
});
axiosRetry(axiosClient, { retries: 3 }); // retry non-POST requests on network or 5XX errors

export function post(url, data, config) {
  return request({
    method: "POST",
    url: url,
    data: data,
    ...config
  });
}

export function put(url, data, config) {
  return request({
    method: "PUT",
    url: url,
    data: data,
    ...config
  });
}

export function get(url, config) {
  return request({
    method: "GET",
    url: url,
    ...config
  });
}

export function del(url, config) {
  return request({
    method: "DELETE",
    url: url,
    ...config
  });
}

async function request(config) {
  let token = await Auth.getAccessToken();
  try {
    return await requestWithToken(config, token);
  } catch (e) {
    // throw non-401 errors
    if (!e.response || e.response.status !== 401) {
      throw e;
    }
    // retry 401 errors one time
    console.log("request retry 401", e.response);
    token = await Auth.refreshAccessToken();
    console.log("request refreshed token", token);
    return await requestWithToken(config, token);
  }
}

function requestWithToken(config, token) {
  return axiosClient.request({
    ...config,
    headers: {
      Authorization: `Bearer ${token}`,
      ...config.headers
    }
  });
}
