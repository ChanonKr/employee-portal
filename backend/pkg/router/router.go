// Package router holds what the server hands to each module so the module can
// register its own routes. It lives outside modules/server to avoid an import
// cycle (server imports every module).
package router

import "github.com/gin-gonic/gin"

type Groups struct {
	// Root has no body limit: for routes that set their own (photo upload).
	Root *gin.RouterGroup
	// API is /api/v1 with the JSON body limit and no authentication.
	API *gin.RouterGroup
	// Admin is /api/v1/admin: signed in, password changed, role admin.
	Admin *gin.RouterGroup

	// SignedIn requires a valid session.
	SignedIn gin.HandlerFunc
	// Ready also rejects sessions still on a temporary password.
	Ready gin.HandlerFunc
}
