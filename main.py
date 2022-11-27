#!/usr/bin/env python3
"""
Find correlations in a food journal.

Found correlations are delayed autocorrelations of qualitative data; i.e.
correlations of the same variable across time, where the variable is a string
description (an "event").

Examples of events are "coffee", "good poop", "bad poop", "exercise", "sugar",
"alcohol", etc.; this program tries to find correlations between "[quality]
poop" and any other value.

The input must be a Numbers.app sheet.
"""

from datetime import datetime, timedelta, timezone

import pandas
from numbers_parser import Document


class Food:
    """A food type."""

    def __init__(self, name: str):
        """Initialize the food from a string."""
        self.name = name


class Poop:
    """A wrapper around a poop row."""

    BAD: str = "bad poop"
    GOOD: str = "good poop"

    def __init__(self, index: int, datetime: datetime, type: str):
        """Initialize the poop from a row."""
        if type.lower() not in {Poop.BAD, Poop.GOOD}:
            raise Exception(f"Invalid poop type: {type}")

        self.index = index
        self.datetime = datetime
        self.type = type


class Row:
    """A row of data in the Numbers sheet."""

    POOP: str = "poop"
    FOOD: str = "food"

    def __init__(self, row: pandas.Series):
        """Initialize the row from a Pandas row."""
        self._row = row
        self.datetime: datetime | None = None

        dt: str | datetime | None = row[0].value
        try:
            ev: str = str(row[1].value).strip().lower()
        except AttributeError as e:
            raise AttributeError(f"Invalid event: {row[1].value}") from e

        # self.index = row[0].value
        if isinstance(dt, datetime):
            self.datetime = dt
        elif isinstance(dt, str):
            try:
                self.datetime = datetime.fromisoformat(dt)
            except TypeError as e:
                raise TypeError(f"Invalid datetime: {dt} (type {dt.__class__}): {e}")

        if self.datetime and self.datetime.tzinfo is None:
            self.datetime = self.datetime.replace(
                tzinfo=timezone(-timedelta(hours=8), "PST")
            )
        self.event = ev
        self.type = self.POOP if ev in {"good poop", "bad poop"} else Row.FOOD

    def to_poop(self) -> Poop:
        """Return the row as a Poop type."""
        return Poop(-1, self.datetime, self.event)


class Cupboard:
    """A collection of food types."""

    def __init__(self) -> None:
        """Initialize the cupboard."""
        self._foods: dict[str, Food] = {}

    def get(self, name: str) -> Food:
        """Get a food from the cupboard."""
        name = name.strip().lower()
        if name not in self._foods:
            self._foods[name] = Food(name)
        return self._foods[name]


def main() -> None:
    """Print delayed autocorrelations in the Numbers sheet."""
    doc = Document(
        "/Users/glacials/Library/Mobile Documents/"
        "com~apple~Numbers/Documents/poops.numbers"
    )
    sheets = doc.sheets
    tables = sheets[0].tables

    cupboard: Cupboard = Cupboard()
    bads: dict[Food, int] = {}
    goods: dict[Food, int] = {}

    data = doc.sheets[0].tables[0].rows(values_only=True)
    df = pandas.DataFrame(data[1:], columns=data[0])

    rows: list[Row] = []
    for row in tables[0].rows()[1:]:
        # TODO: Use data even when no datetime is present.
        if row[1] in {None, ""}:
            continue
        rows.append(Row(row))

    for row in rows:
        if row.type == Row.POOP:
            poop = row.to_poop()
            for food in rows:
                if food.type == Row.POOP:
                    continue
                if food.datetime is None or poop.datetime is None:
                    continue
                if food.datetime > poop.datetime:
                    break
                if poop.datetime - food.datetime < timedelta(hours=24):
                    if poop.type == Poop.BAD:
                        bads[cupboard.get(food.event)] = (
                            bads.get(cupboard.get(food.event), 0) + 1
                        )
                    if poop.type == Poop.GOOD:
                        goods[cupboard.get(food.event)] = (
                            goods.get(cupboard.get(food.event), 0) + 1
                        )

    print("Good foods:")
    for food, num in goods.items():
        if num > 2 and num > bads.get(food, 0):
            print(f"{food.name}: {num}")
    print("")
    print("Bad foods:")
    for food, num in goods.items():
        if num > 2 and num > goods.get(food, 0):
            print(f"{food.name}: {num}")


if __name__ == "__main__":
    main()
