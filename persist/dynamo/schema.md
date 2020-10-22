# DynamoDB Schema

## Current (with string PK)

| Entity              | PK                  | SK / GSI PK        | GSI SK           | Notes
| --                  | --                  | --                 | --               | --
| Category            | ID                  | "category"         | Name             |
| Collection          | ID                  | "collection"       | Name             |
| collection_category | ID                  | "collection_category#" + CategoryID | ID |
| Place               | ID                  | "place#" + First two characters of FullName | FullName |
| PlaceSettings       | "placeSettings"     | "placeSettings"    | \<none\>         |
| PlaceWord           | "placeWord#" + Word | "placeWord"        | \<none\>         |
| Post                | ID                  | "post"             | CollectionID     |
| Record              | ID                  | "record_post#" + PostID | ID        |
| Settings            | "settings"          | "settings"         | Name             | setting name
| Sequence            | "sequence"          | "sequence"         | \<none\>         |
| User                | ID                  | "user"             | SortKey          | SortKey value is URL-encoded Issuer + Subject
| record_household    | PostID              | "record_household#" + Household | PostID           |

## Notes on changes
* Where we don't have a use for the GSI, we don't put values in the GSI SK. In that case there will be no item in the GSI (indicated by _\<none\>_).
* Conversely, when we do have a use for the GSI, the GSI SK must have a value. For _collection_category_, I put _ID_ in the GSI SK to ensure that it is indexed and to provide a stable sort order when querying it.
* For _record_, I made the GSI PK be "record_post#" + PostID. That puts each post's records in a different GSI partition. The downside is that there's no way to query for all records regardless of Post.
* I didn't do the same thing for _post_ WRT _collection_, because we need to be able to query all posts, and we shouldn't have so many posts that a single-partition GSI is an issue.
