/// <reference lib="webworker" />

import { build, files } from "$service-worker";

const worker = self as unknown as ServiceWorkerGlobalScope;
const CACHE_NAME = "my-cache-v1";
const to_cache = build.concat(files);
const staticAssets = new Set(to_cache);

const urlB64ToUint8Array = (base64String) => {
  const padding = "=".repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding)
    .replace(/\-/g, "+")
    .replace(/_/g, "/");
  const rawData = atob(base64);
  const outputArray = new Uint8Array(rawData.length);
  for (let i = 0; i < rawData.length; ++i) {
    outputArray[i] = rawData.charCodeAt(i);
  }
  return outputArray;
};

const saveSubscription = async (subscription) => {
  const SERVER_URL = "http://localhost:8080/save-subscription";
  const response = await fetch(SERVER_URL, {
    method: "post",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(subscription),
  });
  return response.json();
};

worker.addEventListener("install", async () => {
  try {
    const applicationServerKey = urlB64ToUint8Array(
      "BEskS8FtAmwXh88AOPD6T7JXYAyg_1ryvrflshNFOK9BAlqqm85OQ4xXA3FXnCGUOZ14glB0xZk1i6TThmJVVKE"
    );
    const options = { applicationServerKey, userVisibleOnly: true };
    const subscription = await worker.registration.pushManager.subscribe(
      options
    );
    const response = await saveSubscription(subscription);
    console.log(response);
  } catch (err) {
    console.log("Error", err);
  }
});

self.addEventListener("activate", async () => {
  // This will be called only once when the service worker is activated.
  try {
    const applicationServerKey = urlB64ToUint8Array(
      "BEskS8FtAmwXh88AOPD6T7JXYAyg_1ryvrflshNFOK9BAlqqm85OQ4xXA3FXnCGUOZ14glB0xZk1i6TThmJVVKE"
    );
    const options = { applicationServerKey, userVisibleOnly: true };
    const subscription = await worker.registration.pushManager.subscribe(
      options
    );
    console.log("subscribtion succes");
    await saveSubscription(subscription);
    console.log("golang succes");
  } catch (err) {
    console.log("Error", err);
  }
});

self.addEventListener("push", (event) => {
  console.log("[Service Worker] Push Received.");
  console.log(`[Service Worker] Push had this data: "${event.data.text()}"`);

  const title = "Test Webpush";
  const options = {
    body: event.data.text(),
  };

  event.waitUntil(worker.registration.showNotification(title, options));
});

async function fetchAndCache(request) {
  const cache = await caches.open(`offline-${CACHE_NAME}`);

  try {
    const response = await fetch(request);
    cache.put(request, response.clone());
    return response;
  } catch (err) {
    const response = await cache.match(request);
    if (response) return response;
    throw err;
  }
}

worker.addEventListener("fetch", (event) => {
  if (event.request.method !== "GET" || event.request.headers.has("range"))
    return;

  const url = new URL(event.request.url);
  const isHttp = url.protocol.startsWith("http");
  const isDevServerRequest =
    url.hostname === self.location.hostname && url.port !== self.location.port;
  const isStaticAsset =
    url.host === self.location.host && staticAssets.has(url.pathname);
  const skipBecauseUncached =
    event.request.cache === "only-if-cached" && !isStaticAsset;

  if (isHttp && !isDevServerRequest && !skipBecauseUncached) {
    event.respondWith(
      (async () => {
        const cachedAsset =
          isStaticAsset && (await caches.match(event.request));

        return cachedAsset || fetchAndCache(event.request);
      })()
    );
  }
});
