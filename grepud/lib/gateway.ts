import axios from "axios";

const gateway = axios.create({
  baseURL: "http://localhost:8000",
});

gateway.interceptors.request.use((config) => {
  if (typeof window !== "undefined") {
    const token = localStorage.getItem("access_token");
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
});

export default gateway;
