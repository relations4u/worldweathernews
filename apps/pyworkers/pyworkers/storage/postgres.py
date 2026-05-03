"""asyncpg-Pool-Helfer."""

import asyncpg
import structlog

log = structlog.get_logger(__name__)


async def create_pool(database_url: str) -> asyncpg.Pool:
    pool = await asyncpg.create_pool(
        database_url,
        min_size=2,
        max_size=10,
        command_timeout=30,
    )
    if pool is None:
        raise RuntimeError("asyncpg.create_pool returned None")
    log.info("postgres_pool_created")
    return pool


async def health_check(pool: asyncpg.Pool) -> bool:
    try:
        async with pool.acquire() as conn:
            await conn.execute("SELECT 1")
    except Exception as e:
        log.warning("postgres_health_check_failed", error=str(e))
        return False
    return True
