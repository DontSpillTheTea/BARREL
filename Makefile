.PHONY: default help check-env dev up down logs build test test-api test-ocr test-web samples clean status

# Thin compatibility shim delegating to go-task
# If task is missing, show a friendly install hint.

HAS_TASK := $(shell command -v task 2> /dev/null)

define TASK_MISSING_HINT
@echo "BARREL uses go-task as its primary command runner."
@echo "Please install it: https://taskfile.dev/installation/"
@echo "Fallback commands are:"
@echo "  docker compose up --build"
@echo "  docker compose down"
@echo "  python3 scripts/check_env.py"
@echo "  python3 scripts/build_sample_batches.py"
@exit 1
endef

default: help

help:
ifndef HAS_TASK
	$(TASK_MISSING_HINT)
else
	@task help
endif

check-env:
ifndef HAS_TASK
	$(TASK_MISSING_HINT)
else
	@task check-env
endif

dev:
ifndef HAS_TASK
	$(TASK_MISSING_HINT)
else
	@task dev
endif

up:
ifndef HAS_TASK
	$(TASK_MISSING_HINT)
else
	@task up
endif

down:
ifndef HAS_TASK
	$(TASK_MISSING_HINT)
else
	@task down
endif

logs:
ifndef HAS_TASK
	$(TASK_MISSING_HINT)
else
	@task logs
endif

build:
ifndef HAS_TASK
	$(TASK_MISSING_HINT)
else
	@task build
endif

test:
ifndef HAS_TASK
	$(TASK_MISSING_HINT)
else
	@task test
endif

test-api:
ifndef HAS_TASK
	$(TASK_MISSING_HINT)
else
	@task test:api
endif

test-ocr:
ifndef HAS_TASK
	$(TASK_MISSING_HINT)
else
	@task test:ocr
endif

test-web:
ifndef HAS_TASK
	$(TASK_MISSING_HINT)
else
	@task test:web
endif

samples:
ifndef HAS_TASK
	$(TASK_MISSING_HINT)
else
	@task samples
endif

clean:
ifndef HAS_TASK
	$(TASK_MISSING_HINT)
else
	@task clean
endif

status:
ifndef HAS_TASK
	$(TASK_MISSING_HINT)
else
	@task status
endif
