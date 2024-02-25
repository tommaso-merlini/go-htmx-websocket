package handler

import (
	"log/slog"
	"net/http"
	"net/url"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/nedpals/supabase-go"

	"roomate/db"
	"roomate/pkg/sb"
	"roomate/sqlc"
	"roomate/view/auth"
)

func LoginShow(c echo.Context) error {
	return render(c, auth.Login())
}

func Login(c echo.Context) error {
	credentials := supabase.UserCredentials{
		Email:    c.FormValue("email"),
		Password: c.FormValue("password"),
	}
	resp, err := sb.Client.Auth.SignIn(c.Request().Context(), credentials)
	if err != nil {
		slog.Error("login error", "err", err)
		return render(
			c,
			auth.LoginForm(credentials, "The credentials you have entered are invalid"),
		)
	}
	setAccessToken(c, resp.AccessToken)
	http.Redirect(c.Response(), c.Request(), "/", http.StatusSeeOther)
	return nil
}

func RegisterShow(c echo.Context) error {
	if err := setAccessToken(c, "20"); err != nil {
		return err
	}
	return render(c, auth.Register())
}

func Register(c echo.Context) error {
	credentials := auth.RegisterFormProps{
		Email:           c.FormValue("email"),
		Password:        c.FormValue("password"),
		ConfirmPassword: c.FormValue("confirmPassword"),
		Name:            c.FormValue("name"),
	}
	if credentials.Password != credentials.ConfirmPassword {
		return render(
			c,
			auth.RegisterForm(credentials, "Devi inserire la stessa password coglione"),
		)
	}
	u, err := sb.Client.Auth.SignUp(c.Request().Context(), supabase.UserCredentials{
		Email:    credentials.Email,
		Password: credentials.Password,
	})
	if err != nil {
		slog.Error("register error", "err", err)
		return render(
			c,
			auth.RegisterForm(credentials, "The credentials you have entered are invalid"),
		)
	}
	_, err = db.Queries.CreateUser(c.Request().Context(), sqlc.CreateUserParams{
		Email:  credentials.Email,
		Authid: u.ID,
		Name:   credentials.Name,
	})
	if err != nil {
		slog.Error("create user error", "err", err)
		return render(
			c,
			auth.RegisterForm(credentials, "Errore interno"),
		)
	}
	return render(c, auth.SignupSuccess(credentials.Email))
}

func RegisterCallback(c echo.Context) error {
	parsedURL, err := url.Parse(c.Request().URL.String())
	if err != nil {
		return err
	}
	fragment := parsedURL.Fragment
	fragmentValues, err := url.ParseQuery(fragment)
	if err != nil {
		return err
	}
	accessToken := fragmentValues.Get("access_token")
	if err := setAccessToken(c, accessToken); err != nil {
		return err
	}
	http.Redirect(c.Response(), c.Request(), "/", http.StatusSeeOther)
	return nil
}

func setAccessToken(c echo.Context, accessToken string) error {
	session := c.Get("session").(*sessions.Session)
	session.Values[sessionAccessTokenKey] = accessToken
	return session.Save(c.Request(), c.Response())
}
