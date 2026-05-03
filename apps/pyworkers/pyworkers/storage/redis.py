"""Redis-Client-Helfer (asyncio-Variante)."""

import redis.asyncio as redis
import structlog

log = structlog.get_logger(__name__)


async def create_client(redis_url: str) -> redis.Redis:
    client: redis.Redis = redis.from_url(redis_url, decode_responses=True)
    log.info("redis_client_created")
    return client


async def health_check(client: redis.Redis) -> bool:
    try:
        # redis.asyncio.client.ping() ist als Awaitable typisiert; mypy
        # vermutet eine sync-overload und meckert. await ist hier korrekt.
        result: object = await client.ping()  # type: ignore[misc]
        _ = result
    except Exception as e:
        log.warning("redis_health_check_failed", error=str(e))
        return False
    return True
