# Sezzle Calculator

## Overview

## Stack

## Project Structure

## Setup

## Running

## API Reference

## Design Decisions

### Known limitations

- **JSON field names are matched case insensitively.** `{"A": 5, "B": 3}` is
  accepted as `{"a": 5, "b": 3}`. This is `encoding/json` behaviour: struct tags
  are matched exactly first and then case insensitively as a fallback. It is
  inconsistent with `DisallowUnknownFields`, which does reject `{"a":1,"bb":2}`.
  Making it strict would mean decoding into `map[string]json.RawMessage` first
  to validate the keys, and the extra complexity was not judged worthwhile.
- **`Percentage` cannot report an overflow.** It returns a bare `float64`, so a
  result of `+Inf` is caught by a guard in the API layer rather than by the
  domain. The clean fix is to change its signature to `(float64, error)`.
- **A displayed result loses precision on the next operation.** The frontend
  shortens long values for the display, and the next operation is sent using the
  shortened value rather than the exact one.
- **The API contract is declared twice**, in Go structs and in TypeScript types,
  and the two are kept in sync by hand.

## Testing
