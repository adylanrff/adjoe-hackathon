import logging
from typing import Optional, Tuple

import requests

log = logging.getLogger(__name__)

PUBLISHER_BACKEND_URL = "http://localhost:8081"
PUBLISHER_AUTH_TOKEN = "2790c37723d14f8c9964d368e2203325"


def check_balance(cost: int) -> dict:
    """Call the Go publisher backend's /check-balance endpoint."""
    try:
        resp = requests.get(
            f"{PUBLISHER_BACKEND_URL}/check-balance",
            params={"cost": cost},
            headers={"Authorization": PUBLISHER_AUTH_TOKEN},
            timeout=5,
        )
        resp.raise_for_status()
        return resp.json()
    except Exception as e:
        log.warning(f"Balance check failed: {e}")
        return {}


def get_balance_and_campaign(cost: int) -> Tuple[Optional[int], Optional[list], Optional[str]]:
    """Returns (token_balance, offers). Token balance is None on error, offers is None if balance sufficient."""
    data = check_balance(cost)
    tokens = data.get("tokens", 0)
    # If the response contains "Offers", the balance was insufficient (tokens = 0)
    offers = data["Offers"] if "Offers" in data else None
    external_user_id = data["ExternalUserID"] if "ExternalUserID" in data else None

    return tokens, offers, external_user_id
