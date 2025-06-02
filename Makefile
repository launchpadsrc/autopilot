deploy:
	docker buildx build --platform linux/amd64 -t ghcr.io/launchpadsrc/autopilot --push .
