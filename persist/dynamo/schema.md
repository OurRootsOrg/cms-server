# DynamoDB Schema

## Current (with string PK)

| Entity | PK | SK (also GSI PK) | GSI SK Attribute | GSI SK Value
| -- | -- | -- | -- | --
| Category | ID | "category" | \<none\> |
| Collection | ID | "collection" | \<none\> |
| collection_category | ID | "collection_category#" + CategoryID | ID |
| household_record | ID | "household" | PostID |
| Place | ID | "place#" + First two characters of FullName | FullName |
| PlaceSettings | "placeSettings" | "placeSettings" |  |
| PlaceWord | "placeWord#" + Word | "placeWord" |  |
| Post | ID | "post" | CollectionID |
| Record | ID | "record_post#" + PostID | ID |
| Settings | "settings" | "settings" | Name | setting name
| Sequence | "sequence" | "sequence" | \<none\> |
| User | ID | "user" | SortKey | URL-encoded Issuer + Subject

## Notes on changes
* Where we don't have a use for the GSI, we don't put a value in the GSI SK. In that case there will be no item in the GSI (indicated by _\<none\>_).
* Conversely, when we do have a use for the GSI, the GSI SK must have a value. For _collection_category_, I put _ID_ in the GSI SK to ensure that it is indexed and to provide a stable sort order when querying it.
* For _record_, I made the GSI PK be "record_post#" + PostID. That puts each posts records in a different GSI partition. The downside is that there's no way to query for all records regardless of Post.
* I didn't do the same thing for _post_ WRT _collection_, because we need to be able to query all posts, and we shouldn't have so many posts that a single-partition GSI is an issue.
