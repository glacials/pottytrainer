# Pottytrainer

_Note: Pottytrainer is still in its early phases, so is hardcoded to work with my own setup._

Pottytrainer is a bespoke gastrointestinal insights generator. I wrote it because I was having some subpar poops and couldn't trace them to a specific food. There are apps out there that do this, but nothing whose logging process was low-friction enough to keep me from falling off the wagon.

## Architecture

Pottytrainer is composed of the code in this repository and two additional components.

First, to log meals or poops I use iOS shortcuts located on my phone's home screen. This lets the logging process be a single tap to start typing, then one more tap to complete.

The shortcuts log data into the second component, a Numbers.app spreadsheet. The spreadsheet is two columns: a datetime and a string description of the event (e.g. "pizza", "bubble tea", "good poop", "bad poop").

When `main.py` runs, it reads the spreadsheet, does its calculations, and emails me the result.
