# CLI Help Commands - Added Features

## New Help Commands

Added comprehensive help functionality to both `repo` and `user` subcommands.

### Repository Commands Help

```bash
$ gitwapp repo help
Usage: gitwapp repo <command> [args]

Commands:
  add <path>    Add a repository to track
  remove <id>   Remove a repository from tracking
  list          List all tracked repositories
  help          Show this help message
```

Can also use:
- `gitwapp repo -h`
- `gitwapp repo --help`

### User Commands Help

```bash
$ gitwapp user help
Usage: gitwapp user <command> [args]

Commands:
  add <username> <password>   Create a new user
  remove <id>                 Delete a user
  list                        List all users
  passwd <username>           Update user password (not implemented)
  help                        Show this help message
```

Can also use:
- `gitwapp user -h`
- `gitwapp user --help`

## Additional Features

### User List Command
Added `gitwapp user list` command to list all users (previously only repo list was available).

### Improved Error Messages
Unknown commands now suggest running help:
```bash
$ gitwapp repo foo
Error: Unknown repo command: foo

Run 'gitwapp repo help' for usage.
```

## Implementation Details

- Help accessible via `help`, `-h`, or `--help` flags
- Consistent formatting across both subcommands
- User list output shows ID and username (password hash excluded for security)
- Repo list output shows ID, name, and path
