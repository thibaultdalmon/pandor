package main

import "github.com/labstack/echo"

func UserGET(c echo.Context) error {
	mail := c.QueryParam("mail")

	user, err := User.Get(mail)

	if err != nil {
		return c.JSON(404, "")
	}
	// Respond with JSON
	return c.JSON(200, user)
}
func UserPOST(c echo.Context) error {
	return nil
}
func UserPUT(c echo.Context) error {
	return nil
}
func UserDELETE(c echo.Context) error {
	return nil
}
