package client

type AuthClient struct {
	Username string
	Password string
}

func NewAuthClient(username string, password string) *AuthClient {
	return &AuthClient{Username: username, Password: password}
}

// UserAgent contains the user agent used for the push CLI HTTP client
const UserAgent = "PushCLI/0.1 github.com/substitutes/push-cli"
