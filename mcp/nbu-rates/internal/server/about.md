# NBU official exchange rates

- **Source:** the National Bank of Ukraine's public statistics API (`bank.gov.ua`), no key.
- **Meaning:** `rate` is the number of hryvnias (UAH) for **one unit** of the foreign currency on the given
  calendar day. It is a decimal string exactly as published (`"50.5311"`); never treat it as a float.
- **Publication:** the rate for a day is set on the previous banking day. Weekends and holidays carry a
  published rate. A day further ahead than the next published one has no rate yet.
- **Converting:** `amount in UAH = amount × rate`, rounded half up to the kopiyka. Work in minor units
  (cents, kopiykas) and integers.
- **A published rate never changes**, so the server caches it in memory.

## Errors of `get_rate`

A failed call is a tool error whose text starts with a code:

| Code | Meaning | What to do |
|---|---|---|
| `invalid_argument` | the date is not `YYYY-MM-DD`, or the currency is not a three-letter code other than `UAH` | fix the arguments |
| `rate_not_published` | the NBU has no rate of that currency for that day (too far ahead, before the records, or not a quoted currency) | do not guess; ask again later for a future day |
| `upstream_unavailable` | the NBU service did not answer | retry later |
