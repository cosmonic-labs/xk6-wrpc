all:
	@echo "Use 'make infra' or 'make tests'"

infra:
	kubectl apply -k infra

tests:
	kubectl apply -k tests

.PHONY: infra tests
