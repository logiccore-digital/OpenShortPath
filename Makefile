.PHONY: server clean

# Build both dashboard and server
server:
	@echo "Building dashboard..."
	@cd dashboard && npm install && npm run build
	@echo "Copying dashboard build to server directory..."
	@rm -rf server/dashboard-dist
	@mkdir -p server/dashboard-dist
	@cp -r dashboard/dist/* server/dashboard-dist/ 2>/dev/null || true
	@echo "placeholder" > server/dashboard-dist/placeholder.txt
	@echo "Building landing page..."
	@cd landing && npm install && npm run build
	@echo "Copying landing page build to server directory..."
	@rm -rf server/landing-dist
	@mkdir -p server/landing-dist
	@cp -r landing/out/* server/landing-dist/ 2>/dev/null || true
	@echo "placeholder" > server/landing-dist/placeholder.txt
	@echo "Building server..."
	@cd server && go build -o server
	@echo "Build complete!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf dashboard/dist dashboard/node_modules
	@rm -rf landing/out landing/node_modules
	@rm -rf server/dashboard-dist
	@rm -rf server/landing-dist
	@rm -f server/server
	@echo "Clean complete!"

