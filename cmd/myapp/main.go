// @title VPN Management API
// @version 1.0.0
// @description Production-ready WireGuard VPN management microservice with comprehensive API documentation
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.example.com/support
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"github.com/ElissaDesign/vpn-dev/internal/app"
)

func main() {
	app.Start()
}
