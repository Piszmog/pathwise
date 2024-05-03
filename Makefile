format-templ:
	@echo "Formatting templ files..."
	@templ fmt .
generate-templ:
	@echo "Generating templ files..."
	@templ generate -path ./components
generate-templ-watch:
	@echo "Generating templ files..."
	@templ generate -path ./components -watch
generate-tailwind:
	@echo "Generating tailwind files..."
	@tailwindcss -i ./styles/input.css -o ./dist/assets/css/output@dev.css --minify
generate-tailwind-watch:
	@echo "Generating tailwind files..."
	@tailwindcss -i ./styles/input.css -o ./dist/assets/css/output@dev.css --minify --watch
generate-sql:
	@echo "Generating sql files..."
	@sqlc generate
test-e2e:
	@echo "Running E2E tests..."
	@go test ./... -tags=e2e
test-e2e-headful:
	@echo "Running E2E tests..."
	@HEADFUL=true go test ./... -tags=e2e
open-playwright:
	@echo "Opening Playwright..."
	@playwright open
air:
	@echo "Running air..."
	@air
