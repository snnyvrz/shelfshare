.PHONY: dev

dev:
	bunx nx serve books-api

swagger:
	bunx nx run books-api:swagger