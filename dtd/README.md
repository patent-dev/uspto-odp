# USPTO Patent XML DTD References

## ICE (International Common Element) DTD

USPTO uses ICE DTD standards for patent XML documents.

### Patent Grant XML
- **Version**: ICE DTD v4.7 (2022-02-17)
- **Root Element**: `us-patent-grant`
- **DOCTYPE**: `<!DOCTYPE us-patent-grant SYSTEM "us-patent-grant-v47-2022-02-17.dtd">`
- **Namespace**: None (uses DOCTYPE declaration)

### Patent Application XML
- **Version**: ICE DTD v4.6 (2022-02-17)
- **Root Element**: `us-patent-application`
- **DOCTYPE**: `<!DOCTYPE us-patent-application SYSTEM "us-patent-application-v46-2022-02-17.dtd">`
- **Namespace**: None (uses DOCTYPE declaration)

## DTD Availability

USPTO DTD files are embedded in the XML documents via DOCTYPE declarations.
Standalone DTD files are not publicly hosted by USPTO.

The DTD definitions are inferred from the XML structure in bulk data files:
- Patent Grant Bulk Data: https://bulkdata.uspto.gov/data/patent/grant/
- Patent Application Bulk Data: https://bulkdata.uspto.gov/data/patent/application/

## Implementation Notes

The XML parser in this library validates structure based on Go struct tags
matching the ICE DTD element names and attributes. No external DTD validation
is required for normal operation.

Key DTD elements implemented:
- `us-bibliographic-data-grant` / `us-bibliographic-data-application`
- `abstract` with `<p>` paragraphs
- `description` with `<heading>` and `<p>` elements
- `claims` with nested `<claim>` and `<claim-text>` structure
- `document-id` for publication and application references

## Validation

XML structure validation is implicit through Go's XML unmarshaling.
If the XML doesn't match the expected structure, unmarshaling will fail
with descriptive errors.
