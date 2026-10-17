# PIHU WhatsApp MCP

Search conversations and contacts, read messages, and send messages or media.

Standalone optional MCP package for PIHU. Source exported from pihu_mcp and pihu-os.

## Release

1. Review this repository's source, licenses and manifest.
2. Set `MARKET_PUBLISH_TOKEN` in GitHub Actions secrets to the token configured on pihu_web.
3. Set a new manifest version and push the matching `vVERSION` tag.
4. All four native builds must succeed before package files are released and metadata is registered through `https://pihu.nfks.co.in/api/v2`.
5. Published packages require approval in the Market before public installation.

Python archives include a wheelhouse and hashed lock file for offline installation using PIHU's shared Python 3.11 runtime. Node archives include their dependencies and use PIHU's shared Node runtime. Browser uses installed Chrome; browser binaries are not included.

Desktop service integration and account sign-in must be verified before publishing this initial export.
