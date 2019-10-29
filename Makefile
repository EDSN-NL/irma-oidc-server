ALL:
	packr build .
	go build -o irma-oidc-server .
