.PHONY: generate-secret

generate-secret:
	python3 -c "import secrets; print(secrets.token_urlsafe(32))"

up:
	docker compose up -d --build --remove-orphans --force-recreate

down:
	docker compose down