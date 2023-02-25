=== OurRootsDatabase ===
Contributors: dallanq
Tags: genealogy
Requires at least: 5.7
Tested up to: 6.1
Stable tag: 1.0.16
Requires PHP: 7.0
License: GPLv2 or later
License URI: https://www.gnu.org/licenses/gpl-2.0.html

Integrate the OurRoots genealogical records database management system (https://db.ourroots.org) into WordPress.

== Description ==

OurRoots.org provides WordPress-based website hosting for genealogy societies as well as a database management system for genealogical records.
The database management system is open-source and hosted on [GitHub](https://github.com/ourrootsorg/cms-server).
You can either run it yourself, or you can have OurRoots host it for you at https://db.ourroots.org.
This WordPress plugin defines an our-roots short-code that embeds the search client from the database management system into a WordPress-based website.

== Frequently Asked Questions ==

= How can I learn about how to use the OurRoots database management system? =

Videos are available on [YouTube](https://www.youtube.com/channel/UCy2gjiHmtgovMDl0rV4h2VA)

= How can I get support? =
Email dallan at ourroots.org

= How can I modify the search client? =

The javascript code found in the js directory of this plugin is a minified Vue application available in [this Github repository](https://github.com/OurRootsOrg/cms-server).
You can do most modifications by passing various parameters into the short-code or by modifying the CSS without touching the javascript.
Buf if you want to do more - if you want to view or modify the unminified javascript that is embedded via the short-code, read [these instructions](https://github.com/OurRootsOrg/cms-server/blob/master/search-client/README.md).

== Screenshots ==

1. Screenshot showing the database search form embedded in WordPress using this plugin.

== Upgrade Notice ==

No upgrades necessary.

== Changelog ==

= 1.0.9 =
* Prepare to publish plugin on WordPress plugins directory
= 1.0.10 =
* Minor changes to satisfy WordPress plugins directory requirements
= 1.0.11 =
* Sort category/collection facets alphabetically
* New global settings flag to display names as surname, given
* allow span with style tags to pass sanitization
= 1.0.12 =
* Fix faceting bug
= 1.0.13 =
* Previous update didn't include the fix for some reason
= 1.0.14 =
* Update wordpress-tested version to 6.1; include fix from 1.0.12
= 1.0.15 =
* Apply flag to display names as surname, given to record detail page
= 1.0.16 =
* Previous update didn't include the fix