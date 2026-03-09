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
		if errors.Is(err, service.ErrInvalidCredantials) {
			msg = "Invalid email or password"
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
	h.sessions.Put(c.Request.Context(), "flash", "Account created! You can now log in.")
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
