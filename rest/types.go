package rest

import (
	"github.com/tliron/prudence/platform"
)

func RegisterDefaultTypes() {
	platform.RegisterType("Cookie", CreateCookie,
		"name",
		"value",
		"path",
		"domain",
		"expires",
		"maxAge",
		"secure",
		"httpOnly",
		"sameSite",
	)

	platform.RegisterType("Facet", CreateFacet,
		"name",
		"paths",
		"redirectTrailingSlashStatus",
		"variables",
		"representations",
	)

	platform.RegisterType("Representation", CreateRepresentation,
		"name",
		"charSet",
		"redirectTrailingSlash",
		"redirectTrailingSlashStatus",
		"variables",
		"hooks",
		"prepare",
		"describe",
		"present",
		"erase",
		"modify",
		"call",
		"contentTypes",
		"languages",
	)

	platform.RegisterType("Resource", CreateResource,
		"name",
		"variables",
		"facets",
	)

	platform.RegisterType("Route", CreateRoute,
		"name",
		"paths",
		"redirectTrailingSlashStatus",
		"variables",
		"handler",
	)

	platform.RegisterType("Router", CreateRouter,
		"name",
		"variables",
		"routes",
	)

	platform.RegisterType("Server", CreateServer,
		"name",
		"address",
		"port",
		"protocol",
		"tls",
		"ncsaLogFileSuffix",
		"debug",
		"handlerTimeout",
		"readHeaderTimeout",
		"readTimeout",
		"writeTimeout",
		"idleTimeout",
		"handler",
	)

	platform.RegisterType("Static", CreateStatic,
		"root",
		"indexes",
		"presentDirectories",
	)
}
