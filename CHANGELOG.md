# Changelog

## [0.7.0](https://github.com/mc-es/go-auth/compare/v0.6.0...v0.7.0) (2026-02-20)


### ✨ Features

* **domain:** add JWT token management with claims and validation, including tests ([daec37a](https://github.com/mc-es/go-auth/commit/daec37ac5993834351c83700e58ee74e54566727))
* **domain:** add Status and TokenType value objects with validation methods ([f001a00](https://github.com/mc-es/go-auth/commit/f001a00fafc5db670a28c0a58988ed68256ed506))
* **domain:** add Token management with validation and tests ([8225a0d](https://github.com/mc-es/go-auth/commit/8225a0d777513bf924f2d2cbdf94a54da346a152))
* **domain:** add User entity with validation and management methods ([0a97383](https://github.com/mc-es/go-auth/commit/0a97383b8b1936f13283ebd7e4bcf11a4b7b11ec))
* **domain:** implement Email value object with SQL/JSON support and tests ([1e8df7b](https://github.com/mc-es/go-auth/commit/1e8df7bcb5e33ae65ea5e9d19d354279bd876a5d))
* **domain:** implement Password value object with SQL/JSON support and tests ([18e5bb9](https://github.com/mc-es/go-auth/commit/18e5bb970ad8956b741a5cadb2e7fa260cec8c81))
* **domain:** implement PasswordHasher interface and bcrypt hasher with tests ([75a2691](https://github.com/mc-es/go-auth/commit/75a2691e86ae16b6664e379d0329aa43fccfa2c0))
* **domain:** implement Role value object with SQL/JSON support and tests ([387201b](https://github.com/mc-es/go-auth/commit/387201b7e5ef45e6a80517f6ed78131ffa75da10))
* **domain:** implement Session management with validation and tests ([276feab](https://github.com/mc-es/go-auth/commit/276feab7e03ab111eea41f749972cfca61178892))
* **domain:** implement Username value object with SQL/JSON support and tests ([5930950](https://github.com/mc-es/go-auth/commit/593095076fb6e32f9e092acf5e2ad7eb2aa255a7))
* **domain:** introduce UserRepository and SessionRepository interfaces for user and session management ([8fd571b](https://github.com/mc-es/go-auth/commit/8fd571ba173eabb11af1568911a2f3bcabe9d5c5))

## [0.6.0](https://github.com/mc-es/go-auth/compare/v0.5.1...v0.6.0) (2026-02-14)


### ✨ Features

* **makefile:** add migration commands and update Makefile structure for improved database management ([f4cf96e](https://github.com/mc-es/go-auth/commit/f4cf96e25a76e92fe07bd03e8997fd70b32b8b78))
* **migrations:** add initial migration files for users, sessions, and tokens tables with update triggers ([b1fc5ef](https://github.com/mc-es/go-auth/commit/b1fc5ef4896c314b1bd43e5e2317345add78e443))

## [0.5.1](https://github.com/mc-es/go-auth/compare/v0.5.0...v0.5.1) (2026-02-13)


### 🐛 Bug Fixes

* **config:** rename 'auth' to 'security' in configuration files for clarity and consistency ([12ae86f](https://github.com/mc-es/go-auth/commit/12ae86f451b52e81afa4c94c69aa0be8378f26ce))

## [0.5.0](https://github.com/mc-es/go-auth/compare/v0.4.0...v0.5.0) (2026-02-13)


### ✨ Features

* **docker:** add PostgreSQL service with configuration for user, password, database, healthcheck, and networking ([6145afd](https://github.com/mc-es/go-auth/commit/6145afd954e68336c1b49c95402269b26718e1dd))

## [0.4.0](https://github.com/mc-es/go-auth/compare/v0.3.1...v0.4.0) (2026-01-21)


### ✨ Features

* **apperror:** introduce structured error handling with custom error codes and response utilities ([3ad409f](https://github.com/mc-es/go-auth/commit/3ad409f633338d741ab3076931ff7a9165dd7114))


### 🐛 Bug Fixes

* **apperror:** replace New function with newError for improved error handling and update related tests ([7080e5d](https://github.com/mc-es/go-auth/commit/7080e5d2c19779c1b6ff6219cac939683f6b51da))

## [0.3.1](https://github.com/mc-es/go-auth/compare/v0.3.0...v0.3.1) (2026-01-19)


### 🐛 Bug Fixes

* **config:** make logger optional and update related schema, tests and loader ([af4a88d](https://github.com/mc-es/go-auth/commit/af4a88dfe8ba25941438597fff90f69a69af4d45))
* **docker-build:** add .app-config.yml to image, refactor config loader and simplify tests ([f7caa4e](https://github.com/mc-es/go-auth/commit/f7caa4ee5288d1a598a9daf973f8384d563e7824))

## [0.3.0](https://github.com/mc-es/go-auth/compare/v0.2.0...v0.3.0) (2026-01-19)


### ✨ Features

* **config:** introduce configuration management with viper and dotenv support and add logger bootstrap ([aef5444](https://github.com/mc-es/go-auth/commit/aef54442e5834e2e0ac322673d4ae9fb8a0280be))

## [0.2.0](https://github.com/mc-es/go-auth/compare/v0.1.0...v0.2.0) (2026-01-14)


### ✨ Features

* **logger:** add context-aware logging with context extractor and update logger adapters ([8626ffa](https://github.com/mc-es/go-auth/commit/8626ffa07efdfe65cbbf7d41b04fc8473fcb1599))
* **logger:** add logrus adapter for structured logging and update logger configuration ([f80a8ca](https://github.com/mc-es/go-auth/commit/f80a8ca6c003bd144ab1acc11b8a7fe016854e9f))
* **logger:** add NOP logger adapter and corresponding tests for no-operation logging functionality ([c191643](https://github.com/mc-es/go-auth/commit/c191643b7918fe5d7bd4c9378d891da8c8fc8f56))
* **logger:** enhance file rotation with lumberjack configuration and update logger output handling ([cdcb8ab](https://github.com/mc-es/go-auth/commit/cdcb8ab97007a3d8d2389c26dc50697ee024b7d3))
* **logger:** implement Named method for logger adapters and enhance tests with context and child logger support ([4f0becb](https://github.com/mc-es/go-auth/commit/4f0becb45c29147d85738181f260aec504cc232d))
* **logger:** implement structured logging with zap driver and add interface ([afef7d0](https://github.com/mc-es/go-auth/commit/afef7d0bca9f07c2ad16f9ba8a0c5ae51d63649e))
* **logger:** integrate Zerolog adapter for enhanced logging capabilities and add corresponding tests ([e6d97d3](https://github.com/mc-es/go-auth/commit/e6d97d38312ecdf850a92d6b29f16a528417344a))
* **tests:** add support for test watching and benchmarking in Makefile ([b6b2012](https://github.com/mc-es/go-auth/commit/b6b2012950da0b75fce0edead6e3eb39d244e0a8))
* **tests:** add support for test watching and benchmarking in Makefile ([1051d03](https://github.com/mc-es/go-auth/commit/1051d039271c2b71c09b903ac64be1e9574b8c44))


### 🚀 Performance Improvements

* **logger:** update logging methods to use a unified log function and remove redundant helper functions ([6a67041](https://github.com/mc-es/go-auth/commit/6a67041db13bfbaed5d5602455f8d03c27868645))
