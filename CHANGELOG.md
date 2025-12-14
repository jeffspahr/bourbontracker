# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.0.0] - 2024-12-14

### Added
- **Multi-State Support**: Added North Carolina (Wake County) tracking alongside Virginia ABC
- **Smart Caching System**: Intelligent inventory caching based on product listing types
  - "Listed" products update every 24 hours
  - Limited/Allocation/Barrel/Christmas products update hourly
  - 80% reduction in API requests on scheduled runs
- **Listing Type Filtering**: UI filter for NC product listing types (Limited, Allocation, Listed, Barrel, Christmas)
- **Wake County Stores**: Support for all 15 Wake County ABC store locations with geocoding
- **Product Scraper**: NC ABC warehouse scraper for definitive product catalog (3,167 products)

### Changed
- **Conservative Rate Limiting**: Reduced to 3 concurrent requests with 1-second delay to prevent 429 errors
- **Extended Timeout**: Increased deployment timeout from 15 to 60 minutes for full product scans
- **Performance Optimizations**: Tracker now handles 48,850+ items across multiple states

### Fixed
- Missing geolocation for 1222 New Bern Ave store
- Docker build failures due to non-existent map.html reference
- CI workflow intermittent failures

### Performance
- Fresh deployments: ~36 minutes (full scan of all products)
- Scheduled runs: ~3-5 minutes (with smart caching)
- Total inventory: 48,850+ items tracked

## [1.0.1] - Previous Release

### Fixed
- Bug fixes and stability improvements

## [1.0.0] - Initial Release

### Added
- Virginia ABC store tracking
- Map-based inventory visualization
- Cloudflare Pages deployment
- Docker container support

[Unreleased]: https://github.com/jeffspahr/bourbontracker/compare/v2.0.0...HEAD
[2.0.0]: https://github.com/jeffspahr/bourbontracker/compare/v1.0.1...v2.0.0
[1.0.1]: https://github.com/jeffspahr/bourbontracker/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/jeffspahr/bourbontracker/releases/tag/v1.0.0
