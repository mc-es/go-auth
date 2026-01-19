# Changelog

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
