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

import argparse
from datetime import datetime, timedelta, timezone

from email.message import EmailMessage
from numbers_parser import Document
import outgoing

DIGESTION_TIME = timedelta(hours=24)
"""At most how long after food is eaten might it contribute to poop quality?"""

IGNORE_FOODS = {"450.0"}
ALIASES = {
        "grilled cold noodles": "kao leng mian",
        }


class Arguments(argparse.Namespace):
    """Command-line arguments."""
    email: bool = False


class Food:
    """A food type."""

    def __init__(self, name: str):
        """Initialize the food from a string."""
        self.name = name
        self.num_good_poops = 0
        self.num_bad_poops = 0

    def addpoop(self, pooptype: str):
        """Add a poop event that came soon after this food."""
        if pooptype == Poop.GOOD:
            self.num_good_poops += 1
        elif pooptype == Poop.BAD:
            self.num_bad_poops += 1

    def quality(self) -> float:
        """
        Return the quality of poops caused by this food.

        The quality is a float between 0 and 1, where 0 is bad and 1 is good.
        """

        if self.num_bad_poops == 0:
            return 0
        return float(self.num_good_poops) / float(self.num_bad_poops)


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

    class NoTimestampError(Exception):
        """An exception raised when the row has no date."""
        def __init__(self, msg: str):
            self.msg = msg

        def __str__(self):
            return f"No timestamp on row: {self.msg}"

    def __init__(self, row) -> None:
        """Initialize the row from a Numbers row."""
        self._row = row

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
                raise self.NoTimestampError(
                    f"Invalid datetime: {dt} (type {dt.__class__}): {e}",
                )
        else:
            raise self.NoTimestampError(
                f"Invalid datetime: {dt} (type {dt.__class__})"
            )

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
        if name in ALIASES:
            name = ALIASES[name]
        if name not in self._foods:
            self._foods[name] = Food(name)
        return self._foods[name]
    
    def all(self) -> set[Food]:
        """Return all foods in the cupboard."""
        return set(self._foods.values())


def send_email(subject: str, text_body: str, html_body: str) -> None:
    msg = EmailMessage()

# Add the html version.  This converts the message into a multipart/alternative
# container, with the original text message as the first part and the new html
# message as the second part.
    msg["Subject"] = subject
    msg["To"] = "ben@twos.dev"
    msg["From"] = "pottytrain@mainframe.twos.dev"
    msg.set_content(text_body)
    msg.add_alternative( html_body, subtype='html')

    with outgoing.from_config_file("pyproject.toml") as sender:
        # Now send that letter!
        sender.send(msg)


def main(email=False) -> None:
    """Print delayed autocorrelations in the Numbers sheet."""
    doc = Document(
        "/Users/glacials/Library/Mobile Documents/"
        "com~apple~Numbers/Documents/poops.numbers"
    )
    sheets = doc.sheets
    tables = sheets[0].tables

    cupboard: Cupboard = Cupboard()

    rows: list[Row] = []
    for row in tables[0].rows()[1:]:
        # TODO: Use data even when no datetime is present.
        if row[1] in {None, ""}:
            continue
        try:
            rows.append(Row(row))
        except Row.NoTimestampError:
            continue

    for pooprow in rows:
        if pooprow.type != Row.POOP:
            continue
        poop = pooprow.to_poop()
        # For each poop row, go find recent food rows and tag them
        for foodrow in rows:
            if foodrow.type != Row.FOOD or foodrow.event in IGNORE_FOODS:
                continue
            if poop.datetime - foodrow.datetime < timedelta(hours=24):
                cupboard.get(foodrow.event).addpoop(poop.type)
    
    s = ""
    longest_food_name = max(len(food.name) for food in cupboard.all())
    s += f"+ {'-' * longest_food_name}-+---------+\n"
    s += f"| {'food'.rjust(longest_food_name, ' ')} | quality |\n"
    s += f"+ {'-' * longest_food_name}-+---------+\n"
    for food in sorted(list(cupboard.all()), key=lambda f: f.quality()):
        s += (
            f"| {food.name.rjust(longest_food_name, ' ')} |"
            f" {food.quality():.2f}    |\n"
        )
    s += f"+{'-' * longest_food_name}--+---------+\n"

    print(s)
    if email:
        today = datetime.today().strftime('%Y-%m-%d')
        send_email(
            subject=f"Food journal digest, {today}",
            text_body=s,
            html_body=f"<pre style='font: monospace;'>{s}</pre>",
        )


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        prog="pottytrain",
        description="Find out which foods you eat are causing bad poops.",
    )
    parser.add_argument(
        '-e',
        '--email',
        action='store_true',
        help='Send an email with the final results',
    )
    namespace = Arguments()
    args = parser.parse_args(namespace=namespace)
    main(args.email)
