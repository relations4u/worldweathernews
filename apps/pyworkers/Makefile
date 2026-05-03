.PHONY: install dev run test lint typecheck fmt clean

install:
	uv sync

dev:
	uv run watchfiles --filter python "python -m pyworkers" pyworkers

run:
	uv run python -m pyworkers

test:
	uv run pytest

lint:
	uv run ruff check
	uv run ruff format --check

typecheck:
	uv run mypy pyworkers

fmt:
	uv run ruff format
	uv run ruff check --fix

clean:
	rm -rf .venv .pytest_cache .mypy_cache .ruff_cache .coverage htmlcov
