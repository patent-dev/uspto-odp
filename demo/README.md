# USPTO ODP API Client Demo

This interactive demo showcases the USPTO ODP API client library's capabilities for browsing and downloading bulk data files from the USPTO Open Data Portal.

## Running the Demo

```bash
go run demo.go
```

## Example Output


USPTO Bulk Data Download Tool
==============================

Using API key from environment (USPTO_API_KEY)

=== STEP 1: Select a Bulk Data Product ===

Fetching all bulk data products...

Found 40 bulk data products:
----------------------------------------
 1. PTGRAPS         : Patent Grant Full-Text Data (No Images) - APS (YEARLY)
 2. PTGRXML         : Patent Grant Full-Text Data (No Images) - XML (WEEKLY)
 3. APPDT           : Patent Application Full Text Data with Embedded TIFF Images (Application Red Book based on WIPO ST.36) (WEEKLY)
 4. PTGRMP2         : Patent Grant Multi-page PDF Images (WEEKLY)
 5. PTGRDT          : Patent Grant Full Text Data with Embedded TIFF Images (Grant Red Book based on WIPO ST.36) - XML (WEEKLY)
 6. PTBLXML         : Patent Grant Bibliographic (Front Page) Text Data - XML (WEEKLY)
 7. PTGRDSGM        : Patent Grant Full Text Data with Embedded TIFF Images (Grant Red Book based on WIPO ST.36) - SGML (YEARLY)
 8. PTBLAPS         : Patent Grant Bibliographic (Front Page) Text Data - APS (YEARLY)
 9. APPBLXML        : Patent Application Bibliographic (Front Page) Data (WEEKLY)
10. APPXML          : Patent Application Full-Text Data (No Images) (WEEKLY)
11. GZLST           : Patent Official Gazettes (WEEKLY)
12. APPMP2          : Patent Application Multi-Page PDF Images (WEEKLY)
13. TRCFECO2        : Trademark Case File Data for Academia and Researchers (YEARLY)
14. PASDL           : Patent Assignment XML (Ownership) Text - Daily (DAILY)
15. PASYR           : Patent Assignment XML (Ownership) Text - Annual (YEARLY)
16. TRTDXFAP        : Trademark Full Text XML Data (No Images) – Daily Applications (DAILY)
17. TTABTDXF        : Trademark Full Text XML Data (No Images) – Daily TTAB (DAILY)
18. TRTYRAP         : Trademark Full Text XML Data (No Images) – Annual Applications (YEARLY)
19. TRTDXFAG        : Trademark Full Text XML Data (No Images) – Daily Assignments (DAILY)
20. ECOPAIR         : Patent Examination Research Dataset (PatEx) for Academia and Researchers (YEARLY)
21. TRTYRAG         : Trademark Full Text XML Data (No Images) – Annual Assignments (YEARLY)
22. TTABYR          : Trademark Full Text XML Data (No Images) – Annual TTAB (YEARLY)
23. TRASECO         : Trademark Assignment Data for Academia and Researchers (YEARLY)
24. ECORSEXC        : Patent Assignment Data for Academia and Researchers (YEARLY)
25. PTGRSGM         : Patent Grant Full-Text Data (No Images) - SGML (YEARLY)
26. PTBLSGM         : Patent Grant Bibliographic (Front Page) Text Data - SGML (YEARLY)
27. PTLITIG         : Patent Litigation Docket Report Data Files for Academia and Researchers (YEARLY)
28. PTAPPCLM        : Patent and Patent Application Claims Research Dataset for Academia and Researchers (YEARLY)
29. HISTEXC         : Historical Patent Data Files for Academia and Researchers (YEARLY)
30. CPCMCAPP        : Cooperative Patent Classification (CPC) Master Classification Files for U.S. Patent Applications (MONTHLY)
31. PTMNFEE2        : Patent Maintenance Fee Events (WEEKLY)
32. ECOPATAI        : The Artificial Intelligence Patent Dataset (AIPD) for Academia and Researchers (YEARLY)
33. CPCMCPT         : Cooperative Patent Classification (CPC) Master Classification Files for U.S. Patent Grants (MONTHLY)
34. PTAPOATH        : Patent and Patent Application Oath Signature Dataset for Academia and Researchers (YEARLY)
35. PTOFFACT        : Patent Application Office Actions Research Dataset for Academia and Researchers (YEARLY)
36. MOONSHOT        : Cancer Moonshot Patent Data Files for Academia and Researchers (YEARLY)
37. PTFWPRE         : Patent File Wrapper (Bulk Datasets) - Weekly (WEEKLY)
38. PTFWPRD         : Patent File Wrapper (Bulk Datasets) - Daily (DAILY)
39. PEDSXML         : Patent Examination Data System (Bulk Datasets) – XML (YEARLY)
40. PEDSJSON        : Patent Examination Data System (Bulk Datasets) - JSON (YEARLY)

Select product number (or 'q' to quit): 2

=== STEP 2: Browse Files for PTGRXML ===

Fetching files for PTGRXML...

Total files available: 1257

=== Showing files 1-20 of 1257 ===
----------------------------------------
 1. ipg250923.zip (162.07 MB) - 2025-09-23
 2. ipg250916.zip (107.53 MB) - 2025-09-16
 3. ipg250909.zip (138.68 MB) - 2025-09-09
 4. ipg250902.zip (134.53 MB) - 2025-09-02
 5. ipg250826.zip (138.12 MB) - 2025-08-26
 6. ipg250819.zip (143.74 MB) - 2025-08-19
 7. ipg250812.zip (156.29 MB) - 2025-08-12
 8. ipg250805.zip (152.72 MB) - 2025-08-05
 9. ipg250729.zip (163.48 MB) - 2025-07-29
10. ipg250722.zip (127.34 MB) - 2025-07-22
11. ipg250715.zip (163.41 MB) - 2025-07-15
12. ipg250708.zip (161.69 MB) - 2025-07-08
13. ipg250701.zip (150.94 MB) - 2025-07-01
14. ipg250624.zip (132.40 MB) - 2025-06-24
15. ipg250617.zip (165.50 MB) - 2025-06-17
16. ipg250610.zip (96.92 MB) - 2025-06-10
17. ipg250603.zip (146.32 MB) - 2025-06-03
18. ipg250527.zip (160.07 MB) - 2025-05-27
19. ipg250520_r1.zip (156.19 MB) - 2025-05-20
20. ipg250520.zip (157.65 MB) - 2025-05-20

Options:
  1-20: Select file number
  n: Next page
  s: Search for file
  q: Quit

Your choice: 2

=== STEP 3: Download 2025/ipg250916.zip ===
Saving to: ipg250916.zip
Progress: 99.5% | 106.98/107.53 MB | Speed: 16.23 MB/s | ETA: 0s     
Success: downloaded ipg250916.zip
   Size: 107.53 MB
   Time: 6.6 seconds
   Avg Speed: 16.31 MB/s