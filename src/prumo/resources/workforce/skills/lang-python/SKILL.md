---
name: lang-python
description: Python 3.11+ strict typing, mypy --strict, ruff formatting and linting, pytest test suites, and zero bare exceptions.
---

# Python Strict Typing & Quality Contract

## 1. Typing & Tooling
- Target Python 3.11+. Every function and method must have explicit parameter and return type hints.
- Pass `mypy --strict` with zero type ignores.
- Prohibit bare `except:` or `except Exception: pass`. Catch specific exception types.

## 2. Security
- Prohibit `eval()`, `exec()`, and insecure `pickle` loads on untrusted inputs.
