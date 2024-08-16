package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mountgear/helpers"
	"mountgear/models"
	"mountgear/services"
	"mountgear/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Session struct {
	Name           string
	Phone          string
	Email          string
	Password       string
	OTP            string
	OTPExpiry      time.Time
	ReferralCode   string
	ReferredUserID uint
}

// .................................................login page............................................
func GetLoginPage(ctx *gin.Context) {
	helpers.SendResponse(ctx, http.StatusOK, "Welcome to Mountgear  please login ", nil)
}

// .................................................login..................................................
func Login(c *gin.Context) {

	var input models.User
	input.Email = c.PostForm("email")
	input.Password = c.PostForm("password")

	var user models.User

	if err := c.ShouldBind(&input); err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid input", nil)
		return
	}

	if err := models.EmailExists(models.DB, input.Email, &user); err != nil {
		helpers.SendResponse(c, http.StatusUnauthorized, "User not found", nil)
		return
	}

	if !user.IsActive {
		helpers.SendResponse(c, http.StatusUnauthorized, "User is not Active", nil)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		helpers.SendResponse(c, http.StatusUnauthorized, "invalid Password. please enter a valid password", nil)
		return

	}

	tokenString, err := utils.GenerateToken(user.ID) //       changed
	if err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to generate token", nil)
		return
	}

	// c.SetCookie("token", tokenString, 300*72, "/", "localhost", false, true)
	helpers.SendResponse(c, http.StatusOK, "Login Successfull", nil, gin.H{"token": tokenString})

}

// ...............................................signup page..............................................................
func GetSignUpPage(c *gin.Context) {
	helpers.SendResponse(c, http.StatusOK, "Render signup page", nil)
}

// ...................................................signup................................................................
func SignUp(c *gin.Context) {
	name := c.PostForm("name")
	phone := c.PostForm("phone")
	email := c.PostForm("email")
	password := c.PostForm("password")
	referralCode := c.PostForm("referralCode")

	// Validate input
	if err := validateInput(name, phone, email, password); err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid input", nil)
		return
	}

	// Check if email already exists
	if err := models.EmailExists(models.DB, email, nil); err == nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Email already exists", nil)
		return
	}

	// Handle referral code
	refUserID, err := handleReferralCode(referralCode)
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Referral ID not found", nil)
		return
	}

	// Generate hashed password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Failed to process password", nil)
		return
	}

	// Generate OTP
	otp := utils.GenerateOTP()
	log.Printf("%v", otp)

	// Create session
	session := Session{
		Name:           name,
		Phone:          phone,
		Email:          email,
		Password:       string(hashedPassword),
		OTP:            otp,
		OTPExpiry:      time.Now().Add(1 * time.Minute),
		ReferralCode:   referralCode,
		ReferredUserID: refUserID,
	}

	// Store session
	if err := storeSession(c, session); err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Failed to create session", nil)
		return
	}

	// Send OTP
	if err := services.SendVerificationEmail(email, otp); err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Failed to send OTP", nil)
		return
	}

	helpers.SendResponse(c, http.StatusOK, "OTP sent successfully", nil)
}
//..................................................................................................................
func GetOTPVerificationPage(c *gin.Context) {
	helpers.SendResponse(c, http.StatusOK, "Render OTP page", nil)
}

//.................................................................................................
func VerifyOTP(c *gin.Context) {
	session, err := getSession(c)
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid session", nil)
		return
	}

	// Verify OTP
	inputOTP := c.PostForm("otp")
	if inputOTP != session.OTP || time.Now().After(session.OTPExpiry) {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid OTP or OTP has expired", nil)
		return
	}

	numberStr := strconv.FormatUint(uint64(session.ReferredUserID), 10)
	// Create new user
	user := models.User{
		Name:         session.Name,
		Phone:        session.Phone,
		Email:        session.Email,
		Password:     session.Password,
		ReferralCode: utils.GenerateRandomCode(),
		ReferredBy:   numberStr,
		IsActive:     true,
	}

	if err := models.DB.Create(&user).Error; err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to create user", nil)
		return
	}
//.................................................................................................................
	UserID, err := strconv.ParseUint(user.ReferredBy, 10, 64)
	if err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "UserId not created for referral bonus  ", nil)
	}
	// Handle referral bonus
	if UserID > 0 {
		if err := addReferralBonus(uint(UserID)); err != nil {
			log.Printf("Failed to add referral bonus: %v", err)
		}
	}

	// Create wallet for new user
	wallet := models.Wallet{
		UserID:  user.ID,
		Balance: 0,
	}

	if err := models.DB.Create(&wallet).Error; err != nil {
		log.Printf("Failed to create wallet: %v", err)
	}

	helpers.SendResponse(c, http.StatusOK, "OTP verified. User created successfully", nil)
}

//...........................................................................................................................
func addReferralBonus(userID uint) error {
	walletAmount := 100.0

	var wallet models.Wallet
	err := models.DB.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Wallet doesn't exist, create a new one
			newWallet := models.Wallet{
				UserID:  userID,
				Balance: walletAmount,
			}
			return models.DB.Create(&newWallet).Error
		}
		return err
	}

	// Wallet exists, update the balance
	return models.DB.Model(&wallet).Update("balance", gorm.Expr("balance + ?", walletAmount)).Error
}

//............................................................................................................
func ResendOTP(c *gin.Context) {
	session, err := getSession(c)
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid session", nil)
		return
	}

	email := session.Email

	// Generate new OTP
	newOTP := utils.GenerateOTP()

	// Update session with new OTP and expiry time
	session.OTP = newOTP
	session.OTPExpiry = time.Now().Add(1 * time.Minute)

	// Store updated session
	if err := storeSession(c, *session); err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to update session", nil)
		return
	}

	log.Printf("Resending OTP to email: %v", email)
	log.Printf("New OTP: %v", newOTP)

	if err := services.SendVerificationEmail(email, newOTP); err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to send OTP", nil)
		return
	}

	helpers.SendResponse(c, http.StatusOK, "OTP resent successfully", nil)
}


//...................................................................................................................
func GetForgotPasswordPage(c *gin.Context) {
	helpers.SendResponse(c, http.StatusOK, "Forgot Password Page", nil)
}

//.........................................................................................................
func InitiatePasswordReset(c *gin.Context) {
	email := c.PostForm("email")

	var user models.User
	if err := models.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			helpers.SendResponse(c, http.StatusNotFound, "User not found", nil)
		} else {
			helpers.SendResponse(c, http.StatusInternalServerError, "Database error", nil)
		}
		return
	}

	otp := utils.GenerateOTP()
	otpExpiry := time.Now().Add(1 * time.Minute)

	// Create a password reset session
	resetSession := Session{
		Email:     email,
		OTP:       otp,
		OTPExpiry: otpExpiry,
	}

	// Store the reset session
	if err := storeSession(c, resetSession); err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to create reset session", nil)
		return
	}

	log.Printf("Sending password reset OTP to email: %v", email)
	log.Printf("OTP: %v", otp)

	if err := services.SendVerificationEmail(email, otp); err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Failed to send OTP", nil)
		return
	}

	helpers.SendResponse(c, http.StatusOK, "OTP sent successfully", nil)
}
//.........................................................................................................................
func GetResetPasswordPage(c *gin.Context) {
	helpers.SendResponse(c, http.StatusOK, "Reset Password Page, enter E-mail", nil)
}
//..........................................................................................................................
func ResetPassword(c *gin.Context) {
	session, err := getSession(c)
	if err != nil {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid session", nil)
		return
	}

	var input models.User
	var user models.User

	// Verify OTP
	inputOTP := c.PostForm("otp")
	if inputOTP != session.OTP || time.Now().After(session.OTPExpiry) {
		helpers.SendResponse(c, http.StatusBadRequest, "Invalid OTP or OTP has expired", nil)
		return
	}

	input.Email = session.Email
	password := c.PostForm("password")
	confirmPassword := c.PostForm("conf_password")

	if !utils.CheckPasswordComplexity(password) {
		helpers.SendResponse(c, http.StatusBadRequest, "Password is not strong enough", nil)
		return
	}

	if password != confirmPassword {
		helpers.SendResponse(c, http.StatusBadRequest, "Password and Confirm Password do not match", nil)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Password not hashed", nil)
		return
	}

	result := models.DB.Model(&user).Where("email = ?", input.Email).Update("password", string(hashedPassword))
	if result.Error != nil {
		helpers.SendResponse(c, http.StatusInternalServerError, "Error resetting password", nil)
		return
	}

	helpers.SendResponse(c, http.StatusOK, "Password reset successfully", nil)
}

func Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	helpers.SendResponse(c, http.StatusOK, "Logged out successfully", nil)
}

// .......................................................................................................
func storeSession(c *gin.Context, session Session) error {
	// Generate a new session ID
	//	sessionID := uuid.New().String()

	// Marshal the session data to JSON
	sessionData, err := json.Marshal(session)
	if err != nil {
		log.Printf("Failed to marshal session: %v", err)
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	encryptedSessionData, err := utils.Encrypt(sessionData)
	if err != nil {
		log.Printf("Failed to encrypt session data: %v", err)
		return fmt.Errorf("failed to encrypt session data: %w", err)
	}

	secureCookie := gin.Mode() != gin.DebugMode

	//c.SetCookie("session_id", sessionID, int(30*time.Minute.Seconds()), "/", "", secureCookie, true)
	c.SetCookie("session_data", string(encryptedSessionData), int(30*time.Minute.Seconds()), "/", "", secureCookie, true)

	return nil
}
func getSession(c *gin.Context) (*Session, error) {

	sessionData, err := c.Cookie("session_data")
	if err != nil {
		return nil, fmt.Errorf("failed to get session data cookie: %w", err)
	}

	// Decrypt session data (assuming sessionData is a base64 string)
	decryptedSessionData, err := utils.Decrypt(sessionData) // Pass directly as string
	if err != nil {
		log.Printf("Failed to decrypt session data: %v", err)
		return nil, fmt.Errorf("failed to decrypt session data: %w", err)
	}

	var session Session
	if err := json.Unmarshal(decryptedSessionData, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

func validateInput(name, phone, email, password string) error {
	if name == "" || phone == "" || email == "" || password == "" {
		return fmt.Errorf("all fields are required")
	}
	if !utils.EmailValidation(email) {
		return fmt.Errorf("invalid email")
	}
	if !utils.ValidPhoneNumber(phone) {
		return fmt.Errorf("invalid phone number")
	}
	if !utils.CheckPasswordComplexity(password) {
		return fmt.Errorf("password must be at least 4 characters long")
	}
	return nil
}

func handleReferralCode(referralCode string) (uint, error) {
	if referralCode == "" {
		return 0, nil // No referral code provided, which is okay
	}

	var referrer models.User
	err := models.DB.Where("referral_code = ?", referralCode).First(&referrer).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, fmt.Errorf("invalid referral code")
		}
		return 0, fmt.Errorf("error processing referral code: %v", err)
	}

	return referrer.ID, nil
}
