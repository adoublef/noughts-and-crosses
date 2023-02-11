.PHONY: jwt, srv, web

jwt:
	@openssl ecparam -genkey -name prime256v1 -noout -out private/jwt_$(shell date +"%m%d%y%H%M").pem

srv:
	@go run ./cmd/monolith/

web:
	@cd ./web && npm run dev