deploy:
	sqlc generate
	docker buildx build --platform linux/amd64 -t ghcr.io/launchpad-it/autopilot --push .
