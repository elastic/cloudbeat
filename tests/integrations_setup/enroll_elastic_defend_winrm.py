#!/usr/bin/env python
"""
Run elastic-defend-windows.ps1 on a Windows host over WinRM (HTTP, basic auth).
Credentials JSON path: WINDOWS_DEFEND_CREDENTIALS_FILE (see CDR composite action).
"""

import json
import os
import time
from pathlib import Path
from typing import Optional

from loguru import logger

try:
    import winrm
except ImportError as exc:
    logger.error("pywinrm is required: poetry install (see tests/pyproject.toml)")
    raise SystemExit(1) from exc


def _load_creds(path: Path) -> dict:
    with path.open(encoding="utf-8") as f:
        return json.load(f)


def main() -> None:
    """Run elastic-defend-windows.ps1 on the Windows host via WinRM until success or retries exhausted."""
    cred_path = os.getenv("WINDOWS_DEFEND_CREDENTIALS_FILE", "").strip()
    ps1_path = os.getenv("ELASTIC_DEFEND_WINDOWS_PS1", "").strip()
    if not cred_path or not ps1_path:
        logger.error("WINDOWS_DEFEND_CREDENTIALS_FILE and ELASTIC_DEFEND_WINDOWS_PS1 must be set")
        raise SystemExit(1)

    creds = _load_creds(Path(cred_path))
    host = creds["public_ip"]
    port = int(creds.get("winrm_port", 5985))
    use_ssl = bool(creds.get("winrm_use_ssl", False))
    username = creds.get("username", "Administrator")
    password = creds.get("password", "")
    if not password:
        logger.error("Password missing in credentials file")
        raise SystemExit(1)

    script = Path(ps1_path).read_text(encoding="utf-8")

    max_attempts = int(os.getenv("ELASTIC_DEFEND_WINRM_RETRIES", "36"))
    delay_s = int(os.getenv("ELASTIC_DEFEND_WINRM_RETRY_DELAY", "10"))

    transport = "ssl" if use_ssl else "plaintext"
    target = f"{host}:{port}" if not use_ssl else f"https://{host}:{port}/wsman"
    session = winrm.Session(
        target,
        auth=(username, password),
        transport=transport,
        server_cert_validation="ignore",
    )

    last_err: Optional[Exception] = None
    for attempt in range(1, max_attempts + 1):
        try:
            logger.info(f"WinRM exec attempt {attempt}/{max_attempts} on {target}")
            result = session.run_ps(script)
            if result.status_code != 0:
                logger.error(result.std_err.decode("utf-8", errors="replace"))
                raise RuntimeError(f"PowerShell exited with status {result.status_code}")
            logger.info(result.std_out.decode("utf-8", errors="replace"))
            return
        except Exception as exc:  # pylint: disable=broad-exception-caught
            last_err = exc
            logger.warning(f"WinRM attempt failed: {exc}")
            time.sleep(delay_s)

    logger.error(f"WinRM enrollment failed after {max_attempts} attempts: {last_err}")
    raise SystemExit(1) from last_err


if __name__ == "__main__":
    main()
