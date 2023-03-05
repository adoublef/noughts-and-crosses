.PHONY: jwt, up, down, web

jwt:
	@openssl ecparam -genkey -name prime256v1 -noout -out private/jwt_$(shell date +"%m%d%y%H%M").pem

up:
	@docker compose up -d

down:
	@docker compose down --rmi all

web:
	@cd ./web && npm run dev