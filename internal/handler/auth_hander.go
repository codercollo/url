package handler

import (
	"errors"
	"net/http"
	"url/internal/helpers"
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
