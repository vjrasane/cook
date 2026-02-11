# listonic-cli

CLI for managing [Listonic](https://listonic.com) shopping lists.

## Configuration

### Authentication

Set your Listonic credentials as environment variables:

```sh
export LISTONIC_EMAIL=you@example.com
export LISTONIC_PASSWORD=your-password
```

Credentials are only required for the initial login. After authenticating, a token is cached and reused (including automatic refresh), so subsequent invocations work without the env vars set.

### Token cache

Tokens are cached at `$XDG_CACHE_HOME/listonic/token.json` (defaults to `~/.cache/listonic/token.json`). The cache stores an access token, refresh token, and expiry timestamp. Delete this file to force re-authentication.

## Usage

All commands output JSON.

### Lists

```sh
listonic list get                       # get all lists
listonic list get <list>                # get a single list by name or ID
listonic list create <name>             # create a new list
listonic list delete <list>             # delete a list
listonic list update <list> --name <n>  # rename a list
listonic list clear <list> --all        # remove all items from a list
listonic list clear <list> --checked    # remove only checked items
listonic list items <list>              # list items in a list
```

### Items

All item commands require `--list` to specify which list the item belongs to.

```sh
listonic item add <name> --list <list> [--amount <qty>] [--unit <unit>]
listonic item check <item> --list <list>
listonic item update <item> --list <list> [--check] [--uncheck]
listonic item delete <item> --list <list>
```

The `<list>` argument accepts either a list name (case-insensitive) or numeric ID.

## Build

```sh
go build -o listonic .
```

## Release

Releases are built by GitHub Actions when a tag matching `listonic-cli/v*` is pushed:

```sh
git tag listonic-cli/v0.1.0
git push origin listonic-cli/v0.1.0
```
