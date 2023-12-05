import axios from "axios";
import axiosRetry from "axios-retry";

function getAPIBaseURL() {
  if (window.ourroots.adminDomain) {
    return window.ourroots.adminDomain.replace(/\/$/, "") + "/api";
  }
  return process.env.VUE_APP_API_BASE_URL;
}

const axiosClient = axios.create({
  baseURL: getAPIBaseURL(),
  withCredentials: false, // This is the default
  headers: {
    Accept: "application/json",
    "Content-Type": "application/json"
  }
});
axiosRetry(axiosClient, { retries: 3 }); // retry non-POST requests on network or 5XX errors

export function get(url, config) {
  return request({
    method: "GET",
    url: url,
    ...config
  });
}

async function request(config) {
  return await axiosClient.request({
    ...config,
    headers: {
      Authorization: `Bearer ${window.ourroots.jwt}`,
      ...config.headers
    }
  });
}
