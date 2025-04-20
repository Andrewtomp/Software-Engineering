# Makefile for building the frontend, generating backend certs, and running the backend.

# --- Variables ---
# Detect OS for browser opening command
# Default to xdg-open (Linux)
OPEN_CMD = xdg-open
ifeq ($(OS),Windows_NT)
	OPEN_CMD = start
else
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Darwin)
		OPEN_CMD = open
	endif
endif

WAIT_SECONDS = 10

BROWSER_URL := https://localhost:8080/

# --- Targets ---

# Default target: run everything
.PHONY: all
all: run

# Build the frontend application
.PHONY: build-frontend
build-frontend:
	@echo "Building frontend application in front-runner..."
	@cd front-runner && npm install && npm run build
	@echo "Frontend build complete."

# Generate backend TLS certificates
.PHONY: generate-certs
generate-certs:
	@echo "Generating backend certificates in front-runner_backend..."
	@(cd front-runner_backend && bash ./generateCert.sh)
	@echo "Certificate generation complete."
	@# Add a small pause in case cert generation is too fast for file system sync
	@sleep 1

# Run the backend server and open the browser
# Depends on frontend build and cert generation
.PHONY: run
run: build-frontend generate-certs
	@echo "Starting Go backend server (front-runner_backend/main.go) in FOREGROUND..."
	@echo "The server will run here. Press Ctrl+C to stop it."
	@# Run the Go server in the foreground (removed the '&')
	@ cd front-runner_backend && go run .
	@ sleep $(WAIT_SECONDS)
	@ $(OPEN_CMD) $(BROWSER_URL)
	@# --- The lines below will ONLY execute AFTER the Go server stops ---
	@echo "Go backend server stopped."

.PHONY: ngrok
ngrok: build-frontend generate-certs
	@echo "Starting Go backend server (front-runner_backend/main.go) in FOREGROUND..."
	@echo "The server will run here. Press Ctrl+C to stop it."
	@# Run the Go server in the foreground (removed the '&')
	@ cd front-runner_backend && go run . --ngrok
	@# --- The lines below will ONLY execute AFTER the Go server stops ---
	@echo "Go backend server stopped."

# Clean up build artifacts and generated certificates
.PHONY: clean
clean:
	@echo "Cleaning up build artifacts and certificates..."
	@rm -rf front-runner/build front-runner/node_modules
	@# Adjust the certificate file patterns if your script generates different names
	@rm -f front-runner_backend/*.crt front-runner_backend/*.key front-runner_backend/.storefrontkey
	@echo "Cleanup complete."

# Optional: Target to only run the backend (assuming build and certs are done)
.PHONY: run-backend-only
run-backend-only:
	@echo "Starting Go backend server (assuming prerequisites are met)..."
	@(cd front-runner_backend && go run .) & \
	SERVER_PID=$$!; \
	echo "Backend server started (PID: $$SERVER_PID)."; \
	echo "Waiting $(WAIT_SECONDS) seconds before opening browser..."; \
	sleep $(WAIT_SECONDS); \
	echo "Opening $(BROWSER_URL) in your browser..."; \
	$(OPEN_CMD) $(BROWSER_URL); \
	echo "Makefile finished. Backend server is running in the background."