.PHONY: lint
lint: custom-gcl
	./custom-gcl run

custom-gcl: .custom-gcl.yml
	golangci-lint custom -v
