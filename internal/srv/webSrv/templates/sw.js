var cacheName = 'mifasol-pwa-{{.}}';

/* Start the service worker and cache all of the app's content */
self.addEventListener('install', function (e) {
    e.waitUntil(
        caches.open(cacheName).then(function (cache) {
            return cache.addAll(
                [
                    '/',
                    '/static/css/fontawesome.css',
                    '/static/css/normalize.css',
                    '/static/css/solid.css',
                    '/static/css/style.css',
                    '/static/font/Inter-ExtraBold.woff2',
                    '/static/font/Inter-Light.woff2',
                    '/static/font/Inter-Medium.woff2',
                    '/static/js/mifasol.js',
                    '/static/js/wasm_exec.js',
                    '/clients/mifasolcliwa.wasm'
                ]
            );
        })
    );
});

/* Serve cached content when offline */
self.addEventListener('fetch', function (e) {
    e.respondWith(
        caches.match(e.request).then(function (response) {
            return response || fetch(e.request);
        })
    );
});

/* Clean old cache */
self.addEventListener('activate', (e) => {
    e.waitUntil(
        caches.keys().then((keyList) => {
            return Promise.all(keyList.map((key) => {
                if (cacheName.indexOf(key) === -1) {
                    return caches.delete(key);
                }
            }));
        })
    );
});
