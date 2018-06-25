.PHONY: test 
test: ## Run tests for Skycoin
	go test ./client/... -timeout=5m
	go test ./provider/... -timeout=5m
	go test ./tracker/... -timeout=5m
