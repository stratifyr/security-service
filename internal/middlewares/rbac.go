package middlewares

import (
	"net/http"

	"github.com/gorilla/mux"
)

const (
	RoleSuperAdmin = "SUPER_ADMIN"
	RoleAdmin      = "ADMIN"
)

var restrictedAPIs = map[string]string{ //nolint:gochecknoglobals // static route-role mapping
	"POST /securities":       RoleSuperAdmin,
	"PATCH /securities/{id}": RoleSuperAdmin,

	"POST /security-stats":       RoleSuperAdmin,
	"PATCH /security-stats/{id}": RoleSuperAdmin,

	"POST /market-holidays":   RoleAdmin,
	"PATCH /market-holidays":  RoleAdmin,
	"DELETE /market-holidays": RoleAdmin,

	"GET /market-data-jobs":         RoleAdmin,
	"POST /market-data-jobs":        RoleAdmin,
	"GET /market-data-jobs/{id}":    RoleAdmin,
	"PATCH /market-data-jobs/{id}":  RoleAdmin,
	"DELETE /market-data-jobs/{id}": RoleAdmin,
}

func RBAC(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		if route == nil {
			next.ServeHTTP(w, r)
			return
		}

		pattern, err := route.GetPathTemplate()
		if err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		requiredRole, isRestricted := restrictedAPIs[r.Method+" "+pattern]
		if isRestricted && r.Header.Get("X-User-Role") != requiredRole {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
