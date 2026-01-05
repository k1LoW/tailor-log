# Changelog

## [v0.6.0](https://github.com/k1LoW/tailor-log/compare/v0.5.2...v0.6.0) - 2026-01-05
### New Features 🎉
- refactor: add RestoreAt and rename From to At by @k1LoW in https://github.com/k1LoW/tailor-log/pull/50
### Fix bug 🐛
- fix: limit minimum time for stored positions to minTimeOffset (18 hours ago) by @k1LoW in https://github.com/k1LoW/tailor-log/pull/49
### Other Changes
- chore(deps): bump github.com/k1LoW/go-github-client/v75 from 75.0.19 to 75.0.21 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/tailor-log/pull/40
- chore(deps): bump the dependencies group with 2 updates by @dependabot[bot] in https://github.com/k1LoW/tailor-log/pull/39
- chore(deps): bump the dependencies group across 1 directory with 4 updates by @dependabot[bot] in https://github.com/k1LoW/tailor-log/pull/44
- chore(deps): bump the dependencies group across 1 directory with 2 updates by @dependabot[bot] in https://github.com/k1LoW/tailor-log/pull/46
- chore(deps): bump the dependencies group with 4 updates by @dependabot[bot] in https://github.com/k1LoW/tailor-log/pull/45
- chore(deps): bump the dependencies group with 3 updates by @dependabot[bot] in https://github.com/k1LoW/tailor-log/pull/48
- chore(deps): bump k1LoW/gh-setup from 1.11.5 to 1.11.6 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/tailor-log/pull/47

## [v0.5.2](https://github.com/k1LoW/tailor-log/compare/v0.5.1...v0.5.2) - 2025-11-20
### Other Changes
- chore(deps): bump the dependencies group across 1 directory with 2 updates by @dependabot[bot] in https://github.com/k1LoW/tailor-log/pull/36
- chore(deps): bump golang.org/x/crypto from 0.43.0 to 0.45.0 by @dependabot[bot] in https://github.com/k1LoW/tailor-log/pull/38

## [v0.5.1](https://github.com/k1LoW/tailor-log/compare/v0.5.0...v0.5.1) - 2025-11-06
### Other Changes
- chore: setup trivy for license check by @k1LoW in https://github.com/k1LoW/tailor-log/pull/32
- chore(deps): bump github.com/DataDog/datadog-api-client-go/v2 from 2.47.0 to 2.48.0 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/tailor-log/pull/33
- chore(deps): bump github.com/DataDog/datadog-api-client-go/v2 from 2.48.0 to 2.49.0 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/tailor-log/pull/34

## [v0.5.0](https://github.com/k1LoW/tailor-log/compare/v0.4.0...v0.5.0) - 2025-10-21
### New Features 🎉
- feat: add `stream` command to fetch log from Tailor Platform by @k1LoW in https://github.com/k1LoW/tailor-log/pull/28
### Other Changes
- feat: add support for input filtering with wildcard patterns by @k1LoW in https://github.com/k1LoW/tailor-log/pull/30

## [v0.4.0](https://github.com/k1LoW/tailor-log/compare/v0.3.3...v0.4.0) - 2025-10-21
### New Features 🎉
- feat(function): send function result by @jackchuka in https://github.com/k1LoW/tailor-log/pull/27
### Other Changes
- chore(deps): bump the dependencies group with 3 updates by @dependabot[bot] in https://github.com/k1LoW/tailor-log/pull/26

## [v0.3.3](https://github.com/k1LoW/tailor-log/compare/v0.3.2...v0.3.3) - 2025-10-21

## [v0.3.2](https://github.com/k1LoW/tailor-log/compare/v0.3.1...v0.3.2) - 2025-10-21

## [v0.3.1](https://github.com/k1LoW/tailor-log/compare/v0.3.0...v0.3.1) - 2025-10-21

## [v0.3.0](https://github.com/k1LoW/tailor-log/compare/v0.2.7...v0.3.0) - 2025-10-20
### Fix bug 🐛
- fix: use github-script by @k1LoW in https://github.com/k1LoW/tailor-log/pull/20

## [v0.2.7](https://github.com/k1LoW/tailor-log/compare/v0.2.6...v0.2.7) - 2025-10-20
### Fix bug 🐛
- fix: handle large log submissions and improve buffer management by @k1LoW in https://github.com/k1LoW/tailor-log/pull/19
### Other Changes
- fix: remove concurrency limit by @k1LoW in https://github.com/k1LoW/tailor-log/pull/17

## [v0.2.6](https://github.com/k1LoW/tailor-log/compare/v0.2.5...v0.2.6) - 2025-10-20
### Fix bug 🐛
- fix: improve error handling and log processing by @k1LoW in https://github.com/k1LoW/tailor-log/pull/15
- fix: improve log submission with payload size and capacity checks by @k1LoW in https://github.com/k1LoW/tailor-log/pull/16

## [v0.2.5](https://github.com/k1LoW/tailor-log/compare/v0.2.4...v0.2.5) - 2025-10-20
### Fix bug 🐛
- fix: handle context cancellation and improve channel management by @k1LoW in https://github.com/k1LoW/tailor-log/pull/12

## [v0.2.4](https://github.com/k1LoW/tailor-log/compare/v0.2.3...v0.2.4) - 2025-10-20

## [v0.2.3](https://github.com/k1LoW/tailor-log/compare/v0.2.2...v0.2.3) - 2025-10-20
### Fix bug 🐛
- fix(datadog): reset buffer with proper capacity to avoid potential issues by @k1LoW in https://github.com/k1LoW/tailor-log/pull/9

## [v0.2.2](https://github.com/k1LoW/tailor-log/compare/v0.2.1...v0.2.2) - 2025-10-20
### New Features 🎉
- fix: set limit concurrency by @k1LoW in https://github.com/k1LoW/tailor-log/pull/8

## [v0.2.1](https://github.com/k1LoW/tailor-log/compare/v0.2.0...v0.2.1) - 2025-10-20
### Fix bug 🐛
- fix: fix action by @k1LoW in https://github.com/k1LoW/tailor-log/pull/5

## [v0.2.0](https://github.com/k1LoW/tailor-log/compare/v0.1.0...v0.2.0) - 2025-10-20
### New Features 🎉
- feat: GitHub Action by @k1LoW in https://github.com/k1LoW/tailor-log/pull/3

## [v0.1.0](https://github.com/k1LoW/tailor-log/commits/v0.1.0) - 2025-10-20
### Fix bug 🐛
- fix: pos per workspaceID by @k1LoW in https://github.com/k1LoW/tailor-log/pull/1
