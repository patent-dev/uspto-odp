#!/usr/bin/env bash
#
# refresh-fixtures re-captures the committed golden response bodies under
# testdata/strictdecode/ from the live USPTO ODP API. The deterministic
# regression (TestFixtures in decode_examples_test.go) reads ONLY these committed
# copies, so re-running this script does not change test behavior until the
# refreshed bodies are reviewed and committed.
#
# This library has no demo/, so the fixtures are captured here directly with curl,
# one bounded call per endpoint (search/list calls are truncated to a single
# representative record to keep the goldens small and stable).
#
# Requires USPTO_API_KEY in the environment (and, for any TSDR fixtures added
# later, USPTO_TSDR_API_KEY). Run from the module root: make refresh-fixtures
set -euo pipefail

: "${USPTO_API_KEY:?set USPTO_API_KEY before refreshing fixtures}"

BASE="https://api.uspto.gov"
DIR="testdata/strictdecode"
H="X-API-Key: ${USPTO_API_KEY}"
APP="17248024"        # US 11,646,472 B2
FOREIGN_APP="15000001"
TRIAL="PGR2025-00004"
APPEAL="2026001845"
INTERFERENCE="106130"

mkdir -p "$DIR"

# get NAME URL  -- fetch URL into $DIR/NAME.json, with a short courtesy pause.
get() {
  local name="$1" url="$2"
  printf '  %-36s ' "$name"
  curl -sS -H "$H" "$url" -o "$DIR/$name.json" -w 'http=%{http_code} size=%{size_download}\n'
  sleep 1
}

echo "Refreshing testdata/strictdecode fixtures from the live API..."

# Patent application + sub-resources.
get get_patent                       "$BASE/api/v1/patent/applications/$APP"
get get_patent_adjustment            "$BASE/api/v1/patent/applications/$APP/adjustment"
get get_patent_assignment            "$BASE/api/v1/patent/applications/$APP/assignment"
get get_patent_documents             "$BASE/api/v1/patent/applications/$APP/documents"
get get_patent_continuity            "$BASE/api/v1/patent/applications/$APP/continuity"
get get_patent_foreign_priority      "$BASE/api/v1/patent/applications/$FOREIGN_APP/foreign-priority"
get get_patent_transactions          "$BASE/api/v1/patent/applications/$APP/transactions"
get get_patent_attorney              "$BASE/api/v1/patent/applications/$APP/attorney"
get get_patent_associated_documents  "$BASE/api/v1/patent/applications/$APP/associated-documents"
get search_patents                   "$BASE/api/v1/patent/applications/search?q=applicationNumberText:$APP&limit=1"

# Status codes + bulk data.
get get_status_codes                 "$BASE/api/v1/patent/status-codes?limit=3"
get search_bulk_products             "$BASE/api/v1/datasets/products/search?q=patent%20grant&offset=0&limit=1"
get get_bulk_product                 "$BASE/api/v1/datasets/products/PTGRXML?fileDataFromDate=2024-01-01&fileDataToDate=2024-01-15"

# PTAB trials, appeals, interferences.
get search_trial_decisions           "$BASE/api/v1/patent/trials/decisions/search?q=trialNumber:$TRIAL&offset=0&limit=1"
get get_trial_decisions              "$BASE/api/v1/patent/trials/$TRIAL/decisions"
get search_trial_documents           "$BASE/api/v1/patent/trials/documents/search?q=*:*&offset=0&limit=1"
get get_trial_documents              "$BASE/api/v1/patent/trials/$TRIAL/documents"
get search_trial_proceedings         "$BASE/api/v1/patent/trials/proceedings/search?q=*:*&offset=0&limit=1"
get get_trial_proceeding             "$BASE/api/v1/patent/trials/proceedings/$TRIAL"
get search_appeal_decisions          "$BASE/api/v1/patent/appeals/decisions/search?q=appealNumber:$APPEAL&offset=0&limit=1"
get get_appeal_decisions             "$BASE/api/v1/patent/appeals/$APPEAL/decisions"
get search_interference_decisions    "$BASE/api/v1/patent/interferences/decisions/search?q=interferenceNumber:$INTERFERENCE&offset=0&limit=1"
get get_interference_decisions       "$BASE/api/v1/patent/interferences/$INTERFERENCE/decisions"

# Petitions. get_petition_decision is chained off the search result's record id.
get search_petitions                 "$BASE/api/v1/patent/petitions/search?q=revival&offset=0&limit=1"
REC=$(grep -oE '"petitionDecisionRecordIdentifier":"[0-9a-f-]+"' "$DIR/search_petitions.json" | head -1 | sed -E 's/.*:"([0-9a-f-]+)"/\1/')
if [ -n "$REC" ]; then
  get get_petition_decision          "$BASE/api/v1/patent/petitions/$REC"
else
  echo "  get_petition_decision: SKIP (no record id in search result)"
fi

# Office Action DSAPI /fields.
get get_office_action_fields           "$BASE/api/v1/patent/oa/oa_actions/v1/fields"
get get_office_action_citation_fields  "$BASE/api/v1/patent/oa/oa_citations/v2/fields"
get get_office_action_rejection_fields "$BASE/api/v1/patent/oa/oa_rejections/v2/fields"
get get_enriched_citation_fields       "$BASE/api/v1/patent/oa/enriched_cited_reference_metadata/v3/fields"

echo
echo "Done. NOTE: search/list and get_petition_decision fixtures are truncated to a"
echo "single representative record by hand (or via jq) to keep the goldens small;"
echo "review 'git diff $DIR' before committing."
