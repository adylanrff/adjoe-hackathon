import logging
from typing import Optional

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


def get_campaign_if_needed(cost: int) -> Optional[list]:
    """Returns offers/campaign data if balance is insufficient, None otherwise."""
    data = check_balance(cost)
    # If the response contains "Offers", the balance was insufficient
    if "Offers" in data and data["Offers"]:
        return data["Offers"]
    return None
