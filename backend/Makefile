run:
	go run cmd/app/main.go

migrate:
	atlas migrate hash && \
	atlas migrate apply \
  		--url "sqlite://jellyfin.db" \
  		--dir "file://migrations" && \
	atlas migrate apply \
  		--url "sqlite://emby.db" \
  		--dir "file://migrations"

tygo:
	tygo generate --config tygo.yaml