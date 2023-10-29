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
	@tailwindcss -i ./styles/input.css -o ./assets/css/output.css --minify
generate-tailwind-watch:
	@echo "Generating tailwind files..."
	@tailwindcss -i ./styles/input.css -o ./assets/css/output.css --minify --watch
air:
	@echo "Running air..."
	@air
