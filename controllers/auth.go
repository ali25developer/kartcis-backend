package controllers

import (
	"net/http"
	"os"
	"time"

	"kartcis-backend/config"
	"kartcis-backend/models"

	"crypto/rand"
	"encoding/hex"
	"kartcis-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Phone    string `json:"phone"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	// Check if email exists
	var existingUser models.User
	if err := config.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"success": false, "message": "Email already exists"})
		return
	}

	// Hash password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPassword),
		Phone:    input.Phone,
		Role:     "user",
	}

	config.DB.Create(&user)

	// Generate Token
	token, _ := generateToken(user)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"user":       user,
			"token":      token,
			"expires_in": 7200, // 2 hours
		},
	})
}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid credentials"})
		return
	}

	token, _ := generateToken(user)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user":       user,
			"token":      token,
			"expires_in": 7200,
		},
	})
}

func GetMe(c *gin.Context) {
	// Usually middleware sets "userID" in context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}

func Logout(c *gin.Context) {
	// Stateless JWT, just return success
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Logout successful"})
}

func generateToken(user models.User) (string, error) {
	// In production use proper secret from env
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "secret123"
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 2).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// Get Connected Social Accounts
func GetSocialAccounts(c *gin.Context) {
	userID, _ := c.Get("userID")
	var accounts []models.SocialAccount
	config.DB.Where("user_id = ?", userID).Find(&accounts)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": accounts})
}

// Unlink Social Account
func UnlinkSocialAccount(c *gin.Context) {
	userID, _ := c.Get("userID")
	provider := c.Param("provider")

	// Ensure user has password or at least one other social account before unlinking?
	// For now simple unlink.

	if err := config.DB.Where("user_id = ? AND provider = ?", userID, provider).Delete(&models.SocialAccount{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to unlink account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Account unlinked"})
}

// Set Password (for OAuth users)
func SetPassword(c *gin.Context) {
	userID, _ := c.Get("userID")
	var input struct {
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	config.DB.Model(&models.User{}).Where("id = ?", userID).Update("password", string(hashedPassword))

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Password set successfully"})
}

// Update Profile
func UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("userID") // From middleware

	var input struct {
		Name            string `json:"name"`
		Email           string `json:"email"` // If email update is allowed (might check uniqueness)
		Phone           string `json:"phone"`
		Password        string `json:"password"` // Optional: For password change
		PasswordConfirm string `json:"password_confirm"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "User not found"})
		return
	}

	// Email Uniqueness Check (if email is changing)
	if input.Email != "" && input.Email != user.Email {
		var count int64
		config.DB.Model(&models.User{}).Where("email = ? AND id != ?", input.Email, userID).Count(&count)
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"success": false, "message": "Email already taken"})
			return
		}
		user.Email = input.Email
	}

	if input.Name != "" {
		user.Name = input.Name
	}
	if input.Phone != "" {
		user.Phone = input.Phone
	}

	// Update Password (if provided)
	if input.Password != "" {
		if input.Password != input.PasswordConfirm {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Passwords do not match"})
			return
		}
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		user.Password = string(hashedPassword)
	}

	user.UpdatedAt = time.Now()
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Profile updated successfully", "data": user})
}

// Generate Random Token
func generateRandomString(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Forgot Password
func ForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid email format"})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		// Security: Don't reveal if email exists, just return success
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "If the email is registered, a reset link has been sent."})
		return
	}

	// Generate and Save Token
	token, _ := generateRandomString(32)
	// Wait, standard practice is storing hash in DB, sending plain to user. Or just storing plain.
	// Let's store plain for simplicity in this demo, or bcrypt if you want high security.
	// Let's stick to simple string for easy matching now.

	resetEntry := models.PasswordReset{
		Email:     input.Email,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1 Hour Expiry
		CreatedAt: time.Now(),
	}

	config.DB.Create(&resetEntry)

	// Send Email (Async)
	go utils.SendResetPasswordEmail(user.Email, user.Name, token)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "If the email is registered, a reset link has been sent."})
}

// Reset Password
func ResetPassword(c *gin.Context) {
	var input struct {
		Email           string `json:"email" binding:"required,email"`
		Token           string `json:"token" binding:"required"`
		Password        string `json:"password" binding:"required,min=6"`
		PasswordConfirm string `json:"password_confirm" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	if input.Password != input.PasswordConfirm {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Passwords do not match"})
		return
	}

	// Verify Token
	var resetEntry models.PasswordReset
	// Find latest token for email
	if err := config.DB.Where("email = ? AND token = ?", input.Email, input.Token).Order("created_at desc").First(&resetEntry).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid or expired token"})
		return
	}

	// Check Expiry
	if time.Now().After(resetEntry.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Token has expired, please request a new one"})
		return
	}

	// Update Password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err := config.DB.Model(&models.User{}).Where("email = ?", input.Email).Update("password", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update password"})
		return
	}

	// Invalidate Token (Delete all tokens for this email to be safe)
	config.DB.Where("email = ?", input.Email).Delete(&models.PasswordReset{})

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Password updated successfully. You can now login."})
}
