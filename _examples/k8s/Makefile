all:
	@echo "Use 'make infra' or 'make tests'"

infra:
	kubectl apply -k infra --server-side=true --force-conflicts

tests:
	kubectl apply -k tests

.PHONY: infra tests
