package middleware

import (
 "net/http"
 "strings"

 "github.com/gin-gonic/gin"
 jwtauth "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/auth"
)

func AuthRequired() gin.HandlerFunc {
 return func(c *gin.Context) {
  authHeader := c.GetHeader("Authorization")
  if authHeader == "" {
   c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
   c.Abort()
   return
  }

  parts := strings.Split(authHeader, " ")
  if len(parts) != 2 || parts[0] != "Bearer" {
   c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
   c.Abort()
   return
  }

  claims, err := jwtauth.ValidateToken(parts[1])
  if err != nil {
   c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
   c.Abort()
   return
  }

  c.Set("employee_id", claims.EmployeeID)
  c.Set("email", claims.Email)
  c.Set("role", claims.Role)
  c.Next()
 }
}
