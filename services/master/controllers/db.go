package controllers

import (
	// "context"
	// "log"

	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sdslabs/katana/lib/mongo"
	"github.com/sdslabs/katana/lib/mysql"

	// "github.com/sdslabs/katana/lib/mysql"
	"github.com/sdslabs/katana/lib/utils"
	"github.com/sdslabs/katana/types"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DB(c *fiber.Ctx) error {
	LoadbalancerIP := utils.GetKatanaLoadbalancer()
	mongo.Init(LoadbalancerIP)
	mysql.Init(LoadbalancerIP)
	return c.SendString("Database setup completed\n")
}

func Login(c *fiber.Ctx) error {
	adminUser := new(types.AdminUser)

	if err := c.BodyParser(adminUser); err != nil {
		return err
	}

	admin, err := mongo.FetchSingleAdmin(adminUser.Username)
	if err != nil {
		return err
	}

	isAdmin := utils.CompareHashWithPassword(admin.Password, adminUser.Password)
	if isAdmin {
		jwtToken := jwt.New(jwt.SigningMethodHS256)

		token, err := jwtToken.SignedString([]byte("secret"))
		if err != nil {
			return err
		}

		cookie := new(fiber.Cookie)
		cookie.Name = "jwt"
		cookie.Value = token
		cookie.Expires = time.Now().Add(24 * time.Hour)
		cookie.HTTPOnly = true
		c.Cookie(cookie)

		return c.JSON(fiber.Map{
			"message": "success",
		})
	} else {
		return c.SendString("Incorrect password")
	}
}
