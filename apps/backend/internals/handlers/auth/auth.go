package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	jwtauth "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/auth"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository"
	auth_repo "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository/auth-repo"
)

type AuthHandler struct {
	store *repository.StorageRegistry
}

func NewAuthHandler(store *repository.StorageRegistry) *AuthHandler {
	return &AuthHandler{store: store}
}

type SignupRequest struct {
	Name     string `json:"name"     binding:"required"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type EmployeeResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	Role         string  `json:"role"`
	DepartmentID *string `json:"department_id"`
}

type AuthResponse struct {
	AccessToken  string           `json:"access_token"`
	RefreshToken string           `json:"refresh_token"`
	Employee     EmployeeResponse `json:"employee"`
}

func hashRefreshToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (h *AuthHandler) SignupHandler(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process password"})
		return
	}

	employee, err := h.store.Auth.CreateEmployeeAndUser(c.Request.Context(), req.Name, req.Email, string(hash))
	if err != nil {
		if errors.Is(err, auth_repo.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create account"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "account created successfully",
		"employee_id": employee.ID,
	})
}

func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	employee, err := h.store.Auth.GetEmployeeByEmail(c.Request.Context(), req.Email)
	if err != nil {
		if errors.Is(err, auth_repo.ErrEmployeeNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}

	user, err := h.store.Auth.GetUserByEmployeeID(c.Request.Context(), employee.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	accessToken, err := jwtauth.GenerateAccessToken(employee.ID, employee.Email, employee.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	refreshToken, err := jwtauth.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate refresh token"})
		return
	}

	refreshHash := hashRefreshToken(refreshToken)

	if err := h.store.Auth.StoreRefreshToken(c.Request.Context(), user.ID, refreshHash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store refresh token"})
		return
	}

	if err := h.store.Auth.UpdateLastLogin(c.Request.Context(), user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update login time"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Employee: EmployeeResponse{
			ID:           employee.ID,
			Name:         employee.Name,
			Email:        employee.Email,
			Role:         employee.Role,
			DepartmentID: employee.DepartmentID,
		},
	})
}

func (h *AuthHandler) RefreshTokenHandler(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refreshHash := hashRefreshToken(req.RefreshToken)

	user, err := h.store.Auth.GetUserByRefreshToken(c.Request.Context(), refreshHash)
	if err != nil {
		if errors.Is(err, auth_repo.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate refresh token"})
		return
	}

	employee, err := h.store.Auth.GetEmployeeByID(c.Request.Context(), user.EmployeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get employee"})
		return
	}

	accessToken, err := jwtauth.GenerateAccessToken(employee.ID, employee.Email, employee.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	newRefreshToken, err := jwtauth.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate refresh token"})
		return
	}

	newRefreshHash := hashRefreshToken(newRefreshToken)

	if err := h.store.Auth.StoreRefreshToken(c.Request.Context(), user.ID, newRefreshHash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store refresh token"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		Employee: EmployeeResponse{
			ID:           employee.ID,
			Name:         employee.Name,
			Email:        employee.Email,
			Role:         employee.Role,
			DepartmentID: employee.DepartmentID,
		},
	})
}

func (h *AuthHandler) ForgotPasswordHandler(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.store.Auth.GetEmployeeByEmail(c.Request.Context(), req.Email)
	if err != nil {
		if errors.Is(err, auth_repo.ErrEmployeeNotFound) {
			c.JSON(http.StatusOK, gin.H{"message": "if the email exists, a reset link has been sent"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "if the email exists, a reset link has been sent"})
}

func (h *AuthHandler) MeHandler(c *gin.Context) {
	employeeID, _ := c.Get("employee_id")

	employee, err := h.store.Auth.GetEmployeeByID(c.Request.Context(), employeeID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"employee": EmployeeResponse{
			ID:           employee.ID,
			Name:         employee.Name,
			Email:        employee.Email,
			Role:         employee.Role,
			DepartmentID: employee.DepartmentID,
		},
	})
}
