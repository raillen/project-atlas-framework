from typing import Final

def calculate_tax(amount: float, rate: float) -> float:
    if amount < 0 or rate < 0:
        raise ValueError("Values must be non-negative")
    return amount * rate
