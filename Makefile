.PHONY: default help

# Thin legacy shim delegating to go-task
# Task is the supported command interface. Docker Compose is wrapped by Task.
# Normal reviewers should run task commands, not make or docker compose directly.

define TASK_MISSING_HINT
@echo "BARREL uses task as the supported command interface."
@echo "Run: task help"
@exit 1
endef

default: help

%:
	$(TASK_MISSING_HINT)
