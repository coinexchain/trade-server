<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Client Breaking" for breaking CLI commands and REST routes used by end-users.
"API Breaking" for breaking exported APIs used by developers building on SDK.
"State Machine Breaking" for any changes that result in a different AppState given same genesisState and txList.

Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## [Unreleased]

### State Machine Breaking


### API Breaking Changes


### Client Breaking Changes


### Features


### Improvements


### Bug Fixes


## [v0.0.18] - 2019.10.20 

### State Machine Breaking

*   [#4](https://github.com/coinexchain/trade-server/issues/4) Modify the calculation method of order deal price in trade-server   
*   [#5](https://github.com/coinexchain/trade-server/issues/5) Modify the way depth data is updated to prevent REST queries from getting incomplete data

### Client Breaking Changes
*   [#3](https://github.com/coinexchain/trade-server/issues/3) Add the function of push full amount of depth information periodically
*   [#4](https://github.com/coinexchain/trade-server/issues/4) Modify the calculation method of order deal price in trade-server
*   [#6](https://github.com/coinexchain/trade-server/issues/6) Change LockedCoin type to keep consistent with cetd

