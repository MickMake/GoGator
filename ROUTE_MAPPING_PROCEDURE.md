# GoGator Route Mapping Procedure

## Executive Summary

GoGator should build trips in this order:

```text
raw GPS -> movement/stationary windows -> candidate trips -> site matching -> continuity repair -> route annotation -> output/audit
```

The key idea is to work out where the vehicle actually stopped before deciding what the trip was called. Otherwise the alligator gets to name suburbs, and that way lies dentistry.

---

## 1. Load Inputs

### What it does

Reads:

```text
raw GPS CSV
gogator.yaml
addresses.csv
routes.csv
```

### Why

This establishes the raw tracker data, site definitions, route hints, and matching thresholds before any trip decisions are made.

---

## 2. Parse and Normalise Raw GPS Rows

### What it does

For each raw row, parse:

```text
dt
lat/lng
speed
params
raw row number
```

Apply timezone conversion and timestamp settings.

### Why

Creates a clean internal timeline. Raw GPS remains canonical.

---

## 3. Classify Raw Points

### What it does

Marks each point as broadly:

```text
moving
stationary
uncertain
```

using speed, tracker params, and movement signals.

### Why

This separates actual movement evidence from parked/stopped evidence before building trips.

---

## 4. Build Movement and Dwell Windows

### What it does

Groups neighbouring raw points into windows:

```text
stationary window
movement window
stationary window
movement window
...
```

### Why

Trips should be created between meaningful stops, not from every tiny GPS twitch.

---

## 5. Detect Candidate Trips

### What it does

Creates candidate trips from:

```text
departure dwell window -> movement window -> destination dwell window
```

Each candidate keeps:

```text
raw start/end rows
departure GPS
destination GPS
travel duration
travel distance
tracker evidence
```

### Why

This gives each possible trip enough evidence to be accepted, rejected, merged, or audited.

---

## 6. Suppress Obvious Jitter and Drift

### What it does

Moves non-meaningful fragments to jitter/audit when they are:

```text
too short
same-site movements
GPS drift inside a dwell
stationary teleports
below minimum movement thresholds
```

### Why

The tracker can report real data that is not a meaningful travel-log trip. Real does not always mean useful.

---

## 7. Match Departure and Destination Sites

### What it does

For each candidate trip:

1. Find nearby sites from `addresses.csv`.
2. Check GPS distance against site radius.
3. Use sliding-window dwell evidence to confirm the stop.
4. Choose the best matching site.
5. Otherwise mark as `CHECK`.

### Why

Radius alone is too blunt. Dwell evidence stops the app from labelling a drive-by as a destination, while the sliding window avoids rejecting legitimate sparse tracker pings.

---

## 8. Apply Continuity Repair

### What it does

For adjacent trips, compare:

```text
previous destination site/GPS
current departure site/GPS
```

If the GPS is the same or very close:

```text
CHECK + known site -> repair CHECK to known site
known site + different known site -> flag continuity conflict
CHECK + CHECK -> keep CHECK but audit nearest candidates
```

### Why

A normal trip log should chain cleanly. The destination of one trip is usually the departure of the next. This catches broken labels such as:

```text
Lewis Neal -> CHECK
CHECK -> Geoff Osborne
```

when both `CHECK` points are really Home.

---

## 9. Apply Route Rules

### What it does

Uses `routes.csv` as advisory information:

```text
known route
expected distance range
expected duration range
route confidence
route anomaly
```

### Why

Routes help explain whether a trip looks normal or odd, but they should not silently rewrite destinations.

---

## 10. Recalculate Site Duration

### What it does

After the final trip list is decided, calculate:

```text
site duration = next trip departure time - current trip destination time
```

### Why

Duration must reflect the final cleaned trip chain, not temporary candidate fragments.

---

## 11. Emit Processed Output

### What it does

Writes the main trip table:

```text
Import Index
Continuity/OOPS status
Departure Date/Time
Departure Site
Departure GPS
Departure Address
Travel Duration
Travel Distance
Top Speed
Average Speed
Destination Date/Time
Destination Site
Destination GPS
Destination Address
Site Duration
Flags
```

### Why

This is the spreadsheet-friendly trip log.

---

## 12. Emit Expanded and Audit Outputs

### What it does

Writes supporting files:

```text
expanded raw rows
jitter/rejected trips
route observations
route anomalies
audit report
site-match audit
```

Audit should include things like:

```text
nearest candidate site
distance metres
radius metres
dwell window result
continuity repair reason
CHECK reason
```

### Why

When a trip looks wrong, the answer should be visible without summoning a debugger in a thunderstorm.

---

## 13. Final Validation Pass

### What it does

Checks for suspicious output:

```text
previous destination != current departure
CHECK inside known-site radius
site duration oddities
large wall-clock span with small travel duration
same-site loops
weekday anomalies for known routine sites
```

### Why

This does not necessarily change the result. It highlights rows that need human review.

---

## Compact Flow

```text
1. Load config/sites/routes/raw GPS
2. Parse timestamps and tracker params
3. Classify points as moving/stationary
4. Build movement and dwell windows
5. Create candidate trips
6. Suppress jitter/drift/same-site fragments
7. Match sites using radius + sliding-window dwell
8. Repair adjacent continuity where safe
9. Annotate with route rules
10. Recalculate durations
11. Write processed output
12. Write jitter/expanded/audit outputs
13. Flag suspicious rows
```

---

## Design Principle

Evidence first, labels second, routes last, goblins never.
