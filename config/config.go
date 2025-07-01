package config

const (
	AuthenticatorUsage        = true                    // If the app uses an JWT authenticator. Values: true/false
	AuthenticatorURL          = "http://localhost:8081" // The address to make calls to get a JWT. Values: Your authenticator address
	AppHost                   = "0.0.0.0"               // The host where the app will run. 0.0.0.0 to all addresses. If 0.0.0.0 is specified, the local IP is used for requests
	AppPort                   = 8881                    // The port where the app will run
	AppOpenOnceRunned         = true                    // An option to open the browser at the application address when the application is launched. Values: true/false
	AppCertificateUsage       = false                   // If the HTTPs protocol will be used or not via certificate/key. Values: true/false
	AppCertificateLocation    = ""                      // The location of the .crt/.pem file
	AppCertificateKeyLocation = ""                      // The location of the .key/.pem file
	AppCSRFTokenUsage         = true                    // If the app will use CSRF tokens, to avoid CSRF attacks. Values: true/false
	AppCSRFCookieSecret       = "my-supersecret-cookie" // A secret for the cookie used for CSRF Token verification
	AppCSRFTokenSecret        = "my-supersecret-token"  // A secret for the token used for CSRF Token verification
	CORSUsage                 = true                    // If the app will use CORS. If true, all requests will pass through CORS verification. Values: true/false)
)

var (
	CORSAllowOrigins = []string{"*"} // A list with all the origins allowed.
)
