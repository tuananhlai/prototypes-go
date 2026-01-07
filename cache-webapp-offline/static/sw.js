const CACHE = "offline-v1";
const OFFLINE_URL = "/offline.html";

// Install: cache the offline page
self.addEventListener("install", (event) => {
  event.waitUntil(
    caches.open(CACHE).then((cache) => cache.addAll([OFFLINE_URL])),
  );
  self.skipWaiting();
});

self.addEventListener("activate", (event) => {
  event.waitUntil(self.clients.claim());
});

// Fetch: try network first; if it fails, show offline page for navigations
self.addEventListener("fetch", (event) => {
  const req = event.request;

  // Only handle full page navigations (refresh, address bar, link clicks)
  if (req.mode === "navigate") {
    event.respondWith(fetch(req).catch(() => caches.match(OFFLINE_URL)));
    return;
  }

  // For other requests, just pass through (minimal)
});
