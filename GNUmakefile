default: testacc

# Run acceptance tests. These never touch a live Grafana Cloud tenant - see
# internal/provider/common_test.go for the in-memory fake backend they run
# against.
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m
