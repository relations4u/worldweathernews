"""Click-CLI-Stub. Wird in späteren Sessions für ad-hoc-Kommandos ausgebaut.

Aktuell ohne Subkommandos — der Default-Einstiegspunkt ist ``python -m pyworkers``
(siehe ``__main__.py``).
"""

import click


@click.group()
def cli() -> None:
    """wwn-pyworkers — administrative Kommandos."""


if __name__ == "__main__":
    cli()
