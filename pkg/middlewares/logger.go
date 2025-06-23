package middlewares

// func LoggingMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		tenantID := r.Header.Get("X-Tenant-ID")
// 		log.Infof("Incoming request: %s %s | Tenant: %s", r.Method, r.URL.Path, tenantID)
// 		next.ServeHTTP(w, r)
// 	})
// }
