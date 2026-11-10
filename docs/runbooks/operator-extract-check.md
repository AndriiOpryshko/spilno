# Runbook: the оператор checks the signed extract

For the оператор (the account whose address is `SPILNO_OPERATOR_EMAIL`). Decisions behind it:
association-onboarding [ADR-0005](../features/association-onboarding/adr/0005-check-the-signed-extract-manually-and-keep-the-uploaded-file-in-identity.md)
(the check is manual), [ADR-0009](../features/association-onboarding/adr/0009-treat-the-operator-role-and-a-roster-membership-as-independent-grants-on-one-account.md)
(your own filing). The system never reads the signature; you check it and record the result.

Status: written against the operations of `services/core/api/openapi.yaml`. The decision operations of
tasks T25, T26, T28, T29 and T32 were not merged when this was written, and this procedure has not been
run end to end. Re-check it once they are.

## 1. Open the application

Sign in and open the queue (`GET /identity/operator/applications`, the оператор section of the web app).
Open the заявка на підключення or запит на передачу. Download the applicant's extract of the latest filing
(`GET /identity/operator/applications/{application_id}/filings/{filing_id}/extract`). It arrives as an
attachment; open it from disk, not in the browser.

## 2. Check the КЕП

Open the file in an external tool that validates a qualified electronic signature (КЕП), for example the
verifier of the accredited trust service provider or the state ЦЗО verification service. Use only
a service you trust; do not upload the extract anywhere else.

## 3. Record the check

Record it (`POST /identity/operator/applications/{application_id}/checks`). Write down exactly:

- `signature_valid` - whether the signature is valid;
- `signer_full_name` and `signed_on` - who signed and the signing date, from the КЕП;
- `head_full_name` - the керівник named in the extract;
- `extract_edrpou` - the ЄДРПОУ in the extract, eight digits;
- `extract_dated_on` - the date the extract was generated.

The system then shows warnings. A розбіжність імен: the applicant's name, the signer and the керівник are not
the same person (case, extra spaces, apostrophe form and word order are ignored; another surname, a missing
по батькові or Latin letters is a difference). An extract older than 30 days. Warnings do not block anything
by themselves.

## 4. Decide

- **Names match, code matches:** approve (`.../approval`).
- **Розбіжність імен is open (AC-12):** you may approve only by naming the document relied on, a typed
  reference of up to 500 characters, for example «протокол загальних зборів № 3 від 01.03.2026». No file is
  uploaded. Otherwise the system refuses with `identity.mismatch_needs_document`. You check documents; you do
  not decide who heads the association.
- **The extract has another ЄДРПОУ:** the extract belongs to another association. Approval is refused with
  `identity.extract_code_mismatch`. Reject (`.../rejection`) with a reason the applicant will read, or, if you
  mistyped the code, record the check again with the right one.
- **The signature is invalid or you did not record a check:** approval is refused
  (`identity.signature_not_checked`). Reject with a reason.
- **A rejection always needs a reason** (`identity.reason_required`); the applicant sees it.

## 5. Your own filing

The оператор decides nothing they filed. Approve and reject answer `identity.operator_own_application`; the
check is not refused. Ask another оператор. The guard compares accounts: a second sign-in account of the same
person is not recognised (a known limit, SAD §11).

## Disabling and enabling

`POST /identity/operator/associations/{association_id}/disabling` needs a reason; `.../enabling` restores it.
A вимкнене ОСББ stays readable in the registry and the signed-copy download and refuses everything else.
