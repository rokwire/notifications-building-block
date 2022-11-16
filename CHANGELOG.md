# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased
### Fixed
- Fix docs path and the Dockerfile [#104](https://github.com/rokwire/notifications-building-block/issues/104)

## [1.3.0] - 2022-11-10
### Added
- API for retrieving the count of the unread messages [#95](https://github.com/rokwire/notifications-building-block/issues/95)
- API for filtering "muted" and "unread" [#96](https://github.com/rokwire/notifications-building-block/issues/96)
- API for marking a message as "read" [#94](https://github.com/rokwire/notifications-building-block/issues/94)
- Add a new flag for skipping FCM push notification on creating a new message [#92](https://github.com/rokwire/notifications-building-block/issues/92)
- Support multi-tenancy [#76](https://github.com/rokwire/notifications-building-block/issues/76)

## Fixed
- Fix the docs [#85](https://github.com/rokwire/notifications-building-block/issues/85)

## [1.2.0] - 2022-08-17
### Added
- Async messages capabilities [#83](https://github.com/rokwire/notifications-building-block/issues/83)

## [1.1.8] - 2022-07-15
### Added
- Prepare the project to become open source [#71](https://github.com/rokwire/notifications-building-block/issues/71)
### Security
- Prevent standard users from sending notifications [#77](https://github.com/rokwire/notifications-building-block/issues/77)

## [1.1.7] - 2022-06-08
### Added
- Add support for sending emails as an internal API requested by other building blocks [#72](https://github.com/rokwire/notifications-building-block/issues/72)

## [1.1.6] - 2022-04-28
### Changed
- Update Core auth library to the latest version [#69](https://github.com/rokwire/notifications-building-block/issues/69)

## [1.1.5] - 2022-04-26
### Security
- Update Swagger library due to security issue [#67](https://github.com/rokwire/notifications-building-block/issues/67)

## [1.1.4] - 2022-02-04
### Added
- Put more logs for creating notification [#65](https://github.com/rokwire/notifications-building-block/issues/65)

## [1.1.3] - 2022-01-14
### Fixed
- Fix bad log with wrong error on send Firebase notification  & fix api key log [#63](https://github.com/rokwire/notifications-building-block/issues/63)

## [1.1.2] - 2021-12-02
### Fixed
- Handle all input for recipients, topic and recipient criteria list - do not ignore any of them [#59](https://github.com/rokwire/notifications-building-block/issues/59).

## [0.1.26] - 2021-12-01
### Added
- Implement DELETE /user API for cleaning user info [#57](https://github.com/rokwire/notifications-building-block/issues/57).

## [0.1.25] - 2021-11-23
### Fixed
- Nil pointer error while trying to retrieve missing user [#55](https://github.com/rokwire/notifications-building-block/issues/55).

## [0.1.24] - 2021-11-22
### Fixed
- Message with missing subject must be stored if the body has text [#51](https://github.com/rokwire/notifications-building-block/issues/51).

## [0.1.23] - 2021-11-17
### Added
- Introduce pause notifications for user [#49](https://github.com/rokwire/notifications-building-block/issues/49).

## [0.1.22] - 2021-11-08
### Fixed
- Additional fix for admin API mappings [#46](https://github.com/rokwire/notifications-building-block/issues/46).

## [0.1.21] - 2021-11-05
### Changed
- Update all_admin_notifications permission for accessing admin APIs [#46](https://github.com/rokwire/notifications-building-block/issues/46).

## [0.1.20] - 2021-11-04
### Fixed
- Send FCM to target user only if he/she is subscribed to a topic [#44](https://github.com/rokwire/notifications-building-block/issues/44).

## [0.1.19] - 2021-11-01
### Fixed
- Don't store data notifications without subject & body [#42](https://github.com/rokwire/notifications-building-block/issues/42)


## [0.1.18] - 2021-10-29
### Added
- Expose client API for user record [#40](https://github.com/rokwire/notifications-building-block/issues/40).

## [0.1.17] - 2021-10-27
### Fixed
- Fix bad token transfer from user to anonymous user and vice versa [#37](https://github.com/rokwire/notifications-building-block/issues/37).

## [0.1.16] - 2021-10-27
### Fixed
- Service crash on anonymous token [#35](https://github.com/rokwire/notifications-building-block/issues/35).

## [0.1.15] - 2021-10-27
### Changed
- Improve notifications targeting & filtering [#32](https://github.com/rokwire/notifications-building-block/issues/32).

## [0.1.14] - 2021-10-15
### Fixed
- Messages to topics are not mapped to individual users that are subscribed [#30](https://github.com/rokwire/notifications-building-block/issues/30).

## [0.1.13] - 2021-10-06
### Fixed
- Unable to register new token [#27](https://github.com/rokwire/notifications-building-block/issues/27).

## [0.1.12] - 2021-10-06
### Added
- Еxpose hardcoded config params as environment vars [#25](https://github.com/rokwire/notifications-building-block/issues/25).

## [0.1.11] - 2021-10-05
### Fixed
- Switch from uid to sub identifier from the core token as a user identifier [#23](https://github.com/rokwire/notifications-building-block/issues/23).

## [0.1.10] - 2021-09-29
### Fixed
- Core related fixes [#16](https://github.com/rokwire/notifications-building-block/issues/16).

## [0.1.9] - 2021-09-28
### Added
- Integrate Core BB [#16](https://github.com/rokwire/notifications-building-block/issues/16).

## [0.1.8] - 2021-09-23
### Fixed
-  Fix bad internal message api endpoint and unique topic indexing [#18](https://github.com/rokwire/notifications-building-block/issues/18).

## [0.1.7] - 2021-09-03
### Added
- Introduce Notifications BB [#1](https://github.com/rokwire/notifications-building-block/issues/1).
- Add priority to the message [#5](https://github.com/rokwire/notifications-building-block/issues/5).
- Notifications BB Improvements (Aug/30) [#13](https://github.com/rokwire/notifications-building-block/issues/13).

### Fixed
- Store token & subscribe unsubscribe APIs bug [#14](https://github.com/rokwire/notifications-building-block/issues/14).

## [0.1.0] - 2021-07-19
### Fixed
### Added
