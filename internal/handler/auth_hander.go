package handler

import (
	"errors"
	"net/http"
	"url/internal/helpers"
	"url/internal/mailer"
	"url/internal/service"
	"url/internal/templates"

	"github.com/gin-gonic/gin"
)

// ShowLogin renders the login page
// Pops any flash message left by CreateAccount redirect
func (h *Handler) ShowLogin(c *gin.Context) {
	// PopString reads the value from the session and immediately clears it
	// so the flash only ever shows once
	flash := h.sessions.PopString(c.Request.Context(), "flash")
	helpers.RenderPage(c, "login", &templates.TemplateData{
		Flash: flash,
	})
}

// Login handles admin login form submission
func (h *Handler) Login(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	admin, err := h.auth.Login(c.Request.Context(), email, password)
	if err != nil {
		msg := "Something went wrong, please try again"
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			msg = "Invalid email or password"
		case errors.Is(err, service.ErrAccountInactive):
			msg = "Account not activated - check your email for the activation link"
		}
		helpers.RenderPage(c, "login", &templates.TemplateData{
			Error: msg,
		})
		return
	}

	h.sessions.Put(c.Request.Context(), "adminID", admin.ID)
	h.sessions.Put(c.Request.Context(), "adminUsername", admin.Username)

	// Route must match the registered admin group route
	c.Redirect(http.StatusSeeOther, "/admin/stats")
}

// ShowCreateAccount renders the admin registration page
func (h *Handler) ShowCreateAccount(c *gin.Context) {
	helpers.RenderPage(c, "create_account", &templates.TemplateData{})
}

// CreateAccount handles admin registration form submission
func (h *Handler) CreateAccount(c *gin.Context) {
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")

	err := h.auth.CreateAdmin(c.Request.Context(), username, email, password)
	if err != nil {
		msg := "Something went wrong, please try again"
		switch {
		case errors.Is(err, service.ErrUsernameTaken):
			msg = "Username is already taken"
		case errors.Is(err, service.ErrEmailTaken):
			msg = "An account with that email already exists"
		}
		helpers.RenderPage(c, "create_account", &templates.TemplateData{
			Error: msg,
		})
		return
	}

	// Store flash in session before redirect — ShowLogin will pop it
	h.sessions.Put(c.Request.Context(), "flash", "Account created! Check you email to activate it before signing in.")
	c.Redirect(http.StatusSeeOther, "/login")
}

// ActivateAccount handles the account activation request from a user
func (h *Handler) ActivateAccount(c *gin.Context) {
	//Get the activation token from the query string
	token := c.Query("token")
	if token == "" {
		helpers.RenderPage(c, "login", &templates.TemplateData{
			Error: "Invalid activation link",
		})
		return
	}

	//Call the service layer to activate the account
	err := h.auth.ActivateAccount(c.Request.Context(), token)
	if err != nil {
		msg := "Activation failed. Please try again"

		switch {
		case errors.Is(err, service.ErrTokenInvalid):
			msg = "This activation link is invalid"
		case errors.Is(err, service.ErrTokenExpired):
			msg = "This activation link has expired. Please register again"
		}

		helpers.RenderPage(c, "login", &templates.TemplateData{
			Error: msg,
		})
		return
	}

	//Flash msg when successfull activation
	h.sessions.Put(c.Request.Context(), "flash", "Account activated! You can now sign in")
	c.Redirect(http.StatusSeeOther, "/login")
}

// Logout destroys the session and redirects to the login page
func (h *Handler) Logout(c *gin.Context) {
	if err := h.sessions.Destroy(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not destroy session",
		})
		return
	}
	c.Redirect(http.StatusSeeOther, "/login")
}

// ShowForgotPassword renders the forget password page
func (h *Handler) ShowForgotPassword(c *gin.Context) {
	helpers.RenderPage(c, "forgot_password", &templates.TemplateData{})
}

// ForgotPassword handles the password reset request form
func (h *Handler) ForgotPassword(c *gin.Context) {
	//Get email from submitted form
	email := c.PostForm("email")

	//Generate reset token if the email exists
	token, err := h.auth.ForgotPassword(c.Request.Context(), email)
	if err != nil {
		helpers.RenderPage(c, "forgot_password", &templates.TemplateData{
			Error: "Something went wrong, please try again",
		})
		return
	}

	//If token was generated, enqueue a reset email
	if token != "" {
		resetURL := h.baseURL + "/reset-password?token=" + token

		h.mailer.Enqueue(mailer.Job{
			To:           email,
			Subject:      "Reset your snip.ly password",
			TemplateName: "reset_password.html",
			Data: map[string]string{
				"ResetURL": resetURL,
				"To":       email,
			},
		})
	}

	h.sessions.Put(c.Request.Context(), "flash", "If that email exists, a reset link has been sent.")
	c.Redirect(http.StatusSeeOther, "/login")
}

// ShowResetPassword renders the password reset form using the token
func (h *Handler) ShowResetPassword(c *gin.Context) {
	token := c.Query("token")

	//Redirect if no token is provided
	if token == "" {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	//Render reset password page with token
	helpers.RenderPage(c, "reset_password", &templates.TemplateData{
		Data: map[string]interface{}{
			"token": token,
		},
	})
}

// ResetPassword validates the reset token and updates the password
func (h *Handler) ResetPassword(c *gin.Context) {
	token := c.PostForm("token")
	password := c.PostForm("password")
	confirm := c.PostForm("confirm_password")

	//Ensure passwords match
	if password != confirm {
		helpers.RenderPage(c, "reset_password", &templates.TemplateData{
			Error: "Password do not match",
			Data: map[string]interface{}{
				"token": token,
			},
		})
		return
	}

	//Enforce password length
	if len(password) < 8 {
		helpers.RenderPage(c, "reset_password", &templates.TemplateData{
			Error: "Password must be at least 8 characters ",
			Data: map[string]interface{}{
				"token": token,
			},
		})
		return
	}

	//Attempt password reset
	if err := h.auth.ResetPassword(c.Request.Context(), token, password); err != nil {
		helpers.RenderPage(c, "reset_password", &templates.TemplateData{
			Error: "Reset link is invalid or has expired.",
			Data: map[string]interface{}{
				"token": token,
			},
		})
		return
	}

	//Success message and redirect to login
	h.sessions.Put(c.Request.Context(), "flash", "Password updated. You can now sign in.")
	c.Redirect(http.StatusSeeOther, "/login")

}
