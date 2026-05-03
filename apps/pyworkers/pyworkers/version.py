"""Build-Metadaten — gesetzt via ENV beim Container-Build."""

import os

VERSION: str = os.environ.get("WWN_PY_VERSION", "dev")
COMMIT: str = os.environ.get("WWN_PY_COMMIT", "unknown")
BUILD_DATE: str = os.environ.get("WWN_PY_BUILD_DATE", "unknown")
