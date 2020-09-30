# DynamoDB Schema

## Current (with string PK)

| Entity | PK | SK (also GSI PK) | GSI SK Attribute | GSI SK Value
| -- | -- | -- | -- | --
| Category | ID | "category" | Name |
| Collection | ID | "collection" | Name |
| collection_category | ID | "collection_category#" + CategoryID | n/a | n/a
| household_record | ID | "household" | PostID |
| Place | ID | "place#" + First two characters of FullName | FullName |
| PlaceSettings | "placeSettings" | "placeSettings" |  |
| PlaceWord | "placeWord#" + Word | "placeWord" |  |
| Post | ID | "post" | CollectionID |
| Record | ID | "record" | PostID |
| Settings | "settings" | "settings" | Name | setting name
| Sequence | "sequence" | "sequence" | n/a | n/a
| User | ID | "user" | SortKey | URL-encoded Issuer + Subject
