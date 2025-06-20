run:
	go run cmd/app/main.go

migrate:
	atlas migrate apply \
  		--url "sqlite://$(pwd)/data.db" \
  		--dir "file://migrations"

tygo:
	tygo generate --config tygo.yaml