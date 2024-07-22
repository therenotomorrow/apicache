.PHONY: code docs driver/redis driver/memcached test/smoke test/unit test/integration test/coverage

code:
	@"$(CURDIR)/scripts/code.sh"

docs:
	@"$(CURDIR)/scripts/docs.sh"

driver/redis:
	@"$(CURDIR)/scripts/driver.sh" redis

driver/memcached:
	@"$(CURDIR)/scripts/driver.sh" memcached

test/smoke:
	@"$(CURDIR)/scripts/test.sh" smoke

test/unit:
	@"$(CURDIR)/scripts/test.sh" unit

test/integration:
	@"$(CURDIR)/scripts/test.sh" integration

test/coverage:
	@"$(CURDIR)/scripts/test.sh" coverage
